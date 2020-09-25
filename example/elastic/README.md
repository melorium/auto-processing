# Elastic example

## Filebeat

Tested and verified on Elastic,Kibana,Filebeat 7.9.1.

Filebeat configuration `filebeat.yml` should copied to the filebeat folder example `C:\Program Files\filebeat`

Write in your elastic node address.

Verify the configuration with. `.\filebeat.exe setup -e` and then start with `\filebeat.exe -e`

## Kibana

Go to Stack management, Kibana Saved Objects and then import `kibana.ndjson`

After that go to dashboards and search for automation and you will find the saved object.

 - Automation Dashboard Release 17 ver 2.0

 - Automation Visualization Release 17 ver 2.0

 - Automation Saved Search ERROR Release 17 ver 1.0

 - Automation Saved Search WARN Release 17 ver 1.0

 - Automation Saved Visualization ERROR Release 17 ver 1.0

 - Automation Saved Visualization WARN Release 17 ver 1.0