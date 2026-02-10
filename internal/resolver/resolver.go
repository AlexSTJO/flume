package resolver

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/AlexSTJO/flume/internal/structures"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

var (
	smClient     *secretsmanager.Client
	smClientOnce sync.Once
	smClientErr  error
)

func getSecretsManagerClient() (*secretsmanager.Client, error) {
	smClientOnce.Do(func() {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			smClientErr = fmt.Errorf("loading aws config: %w", err)
			return
		}
		smClient = secretsmanager.NewFromConfig(cfg)
	})
	return smClient, smClientErr
}

var placeholderRE = regexp.MustCompile(`\$\{([^}]+)\}`)

func ResolveString(s string, ctx *structures.Context, infra_outputs *map[string]map[string]string, r *structures.RunInfo) (string, error) {
	var e error
	result := placeholderRE.ReplaceAllStringFunc(s, func(m string) string {
		key := strings.TrimSpace(m[2 : len(m)-1])

		parts := strings.SplitN(key, ":", 2)
		switch parts[0] {
		case "context":
			{
				keys := strings.SplitN(parts[1], ".", 2)
				values := ctx.GetEventValues(keys[0])
				result := values[keys[1]]

				return result
			}
		case "infra":
			keys := strings.SplitN(parts[1], ".", 2)
			return ((*infra_outputs)[keys[0]][keys[1]])
		case "env":
			{
				return os.Getenv(parts[1])
			}
		case "param":
			if r != nil && r.Params != nil {
				if val, ok := r.Params[parts[1]]; ok {
					return val
				}
			}
			e = fmt.Errorf("Unknown parameter: %s", parts[1])
			return "ERROR"
		case "secret":
			secretParts := strings.SplitN(parts[1], ".", 2)
			if len(secretParts) != 2 {
				e = fmt.Errorf("invalid secret reference: %s (expected secret:<provider>.<name>)", key)
				return "ERROR"
			}
			provider := secretParts[0]
			secretName := secretParts[1]

			switch provider {
			case "sm":
				if r != nil && r.SecretsCache != nil {
					if val, ok := r.SecretsCache[secretName]; ok {
						return val
					}
				}

				client, err := getSecretsManagerClient()
				if err != nil {
					e = fmt.Errorf("failed to get Secrets Manager client: %w", err)
					return "ERROR"
				}

				out, err := client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
					SecretId: aws.String(secretName),
				})
				if err != nil {
					e = fmt.Errorf("failed to get secret %s: %w", secretName, err)
					return "ERROR"
				}

				val := aws.ToString(out.SecretString)
				if r != nil && r.SecretsCache != nil {
					r.SecretsCache[secretName] = val
				}
				return val
			default:
				e = fmt.Errorf("unknown secret provider: %s", provider)
				return "ERROR"
			}
		}

		e = fmt.Errorf("Invalid Reference: %s", s)
		return "ERROR"
	})

	if e != nil {
		return "", e
	}

	return result, nil
}

func ResolveStringParam(v string, ctx *structures.Context, infra *map[string]map[string]string, r *structures.RunInfo) (string, error) {
	v, err := ResolveString(v, ctx, infra, r)
	return v, err
}

func ResolveAny(v any, ctx *structures.Context, infra *map[string]map[string]string, r *structures.RunInfo) (any, error) {
	switch typed := v.(type) {
	case string:
		return ResolveString(typed, ctx, infra, r)

	case map[string]any:
		out := make(map[string]any, len(typed))
		for k, val := range typed {
			rv, err := ResolveAny(val, ctx, infra, r)
			if err != nil {
				return nil, err
			}
			out[k] = rv
		}
		return out, nil
	case []any:
		out := make([]any, len(typed))
		for i, val := range typed {
			rv, err := ResolveAny(val, ctx, infra, r)
			if err != nil {
				return nil, err
			}
			out[i] = rv
		}
		return out, nil
	default:
		return v, nil
	}
}

func ToStringSlice(v any) ([]string, error) {
	switch t := v.(type) {
	case []string:
		return t, nil

	case []any:
		out := make([]string, 0, len(t))
		for i, val := range t {
			s, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("attachments[%d] must be a string", i)
			}
			out = append(out, s)
		}
		return out, nil

	default:
		return nil, fmt.Errorf("attachments must be an array of strings")
	}
}
