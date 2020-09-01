# avian-cli Example

* Start the backend-service
* Add remote-servers for remote-connection
* List remote-servers
* Add Nuix Management Servers for licences
* List NM-servers
* List licences
* Add runner for automated workflow
* List runners
* List stages for runners

## Service

The service is our http-server that handles all communication

Start our backend-service (use the flag `-h` or `--help` to see available flags)
```bash
avian service
```

## Handle servers

Add servers to the backend
```bash
avian servers apply `servers.yml`
```

Check out the server in the list
```bash
avian servers list
```

## Handle Nuix Management Servers

Add NMS to the backend
```bash
avian nms apply `nms.yml`
```

List to see our NMS
```bash
avian nms list
```

List our licences for the specified NMS
```bash
avian nms licences `<nms_address>`
```

## Handle the Runners

Add runner to the backend
```bash
avian runners apply `runner.yml`
```

List our runners
```
avian runners list
```

List our stages for the specified Runner
```bash
avian runners stages `<runner_name>`
```
