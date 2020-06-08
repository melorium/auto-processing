## Auto-processing

### Structure overview

```
auto-processing
│   queue.exe
│   auto-processing.exe
|   config.yml 
│
└───configs
│   │   case_1_cfg_example.yml
│   │   case_2_cfg_example.yml
│   
│   
└───scripts
│   └───ruby
│       │   process.rb
│       └───utils
|       |   |    logging.rb
|       |   
|       |
|       └───scripts
|           |    main.rb
|           |    exclude.rb
|           |    ocr.rb
|           |    search_and_tag.rb

```

### queue.exe
Queueing component to run processing-jobs automatically

queue.exe is dependent on a config file, standard path is in the same directory as the queue.exe-file and named config.yml. You can have the config file in another path which is defined as with the cfg-parameter, for example: `queue.exe --cfg=/other/path/cfg.yml`

#### Defining the config-file

In the config-file for the queue-program we only have to define which configs we want to use for the auto-processing.

Example
```yaml
queue:
- config: .\configs\case_1_cfg_example.yml
- config: .\configs\case_2_cfg_example.yml
- config: .\configs\case_3_cfg_example.yml
```

This config is dynamic, which means that we can update it with new auto-processing configs without having to restart the queue.exe. We will also be able to get the status if we open the config. - This is also how the program will now if it should run the auto-processing with the specified config again.

```yaml
queue:
- config: .\configs\case_1_cfg_example.yml
  active: false
  successful: true # Will not run again
  failed: false
- config: .\configs\case_2_cfg_example.yml
  active: false
  successful: true # Will not run again
  failed: false
- config: .\configs\case_3_cfg_example.yml
  active: false
  successful: false
  failed: true # Will try to run again
```

#### Logic

The program will loop through the array of configs, check if it should run or not, after it has looped through the array it will sleep for 5 minutes, then loop again. 