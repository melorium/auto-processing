def ocr(ocr_processor, current_case, search)
    items = current_case.search(search)
    if items.length > 0
        puts("Starting OCR-process for #{items.length} items")
        ocr_processor.process(items)
        puts("OCR-Process finished")
    else
        puts("No items to OCR-process")
    end
end