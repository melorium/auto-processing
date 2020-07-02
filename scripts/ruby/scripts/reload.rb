require File.join(__dir__, "../utils", "logging") 


def reload(processor, current_case, search, current_cfg)
    logger = Logging.get_logger("reload.rb", current_cfg)
    logger.info("START")
    logger.info("Reload from source data started")

    items = current_case.search(search)
    logger.debug("Found #{items.length} items from search: #{search}")
    processor.reload_items_from_source_data(items)
    processor.when_item_processed do |processed_item|
        processed_item_count += 1
        logger.debug("Item reloaded: #{processed_item.get_mime_type()} Count: #{processed_item_count}")
    end
    
    processor.process
    logger.info("Reload from source data finished")
    logger.info("FINISHED")
end