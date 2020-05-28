require File.join(__dir__, "../utils", "logging") 


$logger = Logging.get_logger("ocr.rb")

def ocr(ocr_processor, current_case, search)
    items = current_case.search(search)
    if items.length > 0
        $logger.info("Starting OCR-process for #{items.length} items")
        ocr_processor.process(items)
        $logger.info("OCR-Process finished")
    else
        $logger.debug("No items to OCR-process")
    end
end