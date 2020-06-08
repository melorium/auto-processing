require File.join(__dir__, "../utils", "logging") 


def search_and_tag(current_case, search, tag, search_and_tag_files)
    logger = Logging.get_logger("search_and_tag.rb")
    logger.info("Initializing search and tag")
    if search_and_tag_files.nil?
        logger.debug("Searching through current case: #{search}")
        items = current_case.search(search)
        for item in items
            item.add_tag(tag)
            logger.debug("Tagged item: #{item}")
        end
        logger.info("Tagged #{items.length} items with tag: #{tag}")
        return
    end

    logger.debug("Searching through current case with bulk_searcher")
    # Initialize a bulk searcher with the given scoping query.
    bulk_searcher = current_case.create_bulk_searcher
    # Imports the specified files to the bulk searcher.
    for file in search_and_tag_files
        logger.info("Importing file: #{file} to bulk_searcher")
        bulk_searcher.import_file(file)
    end
    
    logger.info('Performing search and tag...')
    num_rows = bulk_searcher.row_count
    row_num = 0
    # Perform search
    bulk_searcher.run do |progress_info|
        logger.debug("Searching through row #{row_num += 1}/#{num_rows}")
    end
    logger.info("Search and tag finished")
end
