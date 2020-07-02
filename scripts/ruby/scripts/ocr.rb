require File.join(__dir__, "../utils", "logging") 


def ocr(ocr_processor, current_case, search, current_cfg)
    logger = Logging.get_logger("ocr.rb", current_cfg)
    logger.info("START")
    items = current_case.search(search)
    if items.length > 0
        logger.info("Starting OCR-process for #{items.length} items")
        ocr_processor.process(items)
        logger.info("OCR-Process finished")
    else
        logger.debug("No items to OCR-process")
    end
    logger.info("FINISHED")
end