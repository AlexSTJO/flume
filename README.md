# flume

Cool little infra-aware pipeline that I am working on, a bit of a continuation of my cli-flow project


Ok we have a lot of potential solutions for having a server to run this, here is my solution:

Create a command:
    flume start server
    this will read all yaml files and check if there are CRON based messages
    Utilize a post request with the pipeline name and the server will map it to the yaml|
    

/.flume structure
    pipeline_name/
        pipeline_name.yaml
        infra-state.json (probably)


Next direction:
    Attachment handling on smtp
    Need to create when paramater and loop parameter
    Pipe tf error logs into pipeline directory
    potential bug with svc_outputs being overwritten if two of the same services used in steps
