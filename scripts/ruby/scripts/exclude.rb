require File.join(__dir__, "../utils", "logging") 


def exclude(current_case, search, reason, current_cfg)
    logger = Logging.get_logger("exclude.rb", current_cfg)
    logger.info("START")
    items = current_case.search(search)
    for item in items
        item.exclude(reason)
        logger.debug("Excluded item: #{item}")
    end
    logger.info("Excluded #{items.length} items with reason: #{reason}")
    logger.info("FINISHED")
end