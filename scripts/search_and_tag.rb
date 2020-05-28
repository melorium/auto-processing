require File.join(__dir__, "../utils", "logging") 


$logger = Logging.get_logger("search_and_tag.rb")

def search_and_tag(current_case, search, tag)
    items = current_case.search(search)
    for item in items
        item.add_tag(tag)
        $logger.debug("Tagged item: #{item}")
    end
    $logger.info("Tagged #{items.length} items with tag: #{tag}")
end