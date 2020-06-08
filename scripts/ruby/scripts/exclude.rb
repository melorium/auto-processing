require File.join(__dir__, "../utils", "logging") 


def exclude(current_case, search, reason)
    logger = Logging.get_logger("exclude.rb")
    items = current_case.search(search)
    for item in items
        item.exclude(reason)
        logger.debug("Excluded item: #{item}")
    end
    logger.info("Excluded #{items.length} items with reason: #{reason}")
end