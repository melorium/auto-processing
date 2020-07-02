require 'tmpdir'
require 'fileutils'
require File.join(__dir__, "../utils", "logging") 


def populate(current_case, search, current_cfg, types)
    logger = Logging.get_logger("populate.rb", current_cfg)
    logger.info("START")
    logger.info("Populate stores started")

    tmpdir = Dir.tmpdir
    dir = "#{tmpdir}\\populate"
    dir = Dir.mkdir(dir) unless Dir.exist?(dir)

    logger.debug("Creating batch-exporter to temp-dir: #{dir}")
    exporter = $utilities.create_batch_exporter(dir)

    for type in types
        if type == "pdf"
            # Populate pdfs
            exporter.addProduct("pdf",{
                "naming" => "guid",
                "path" => "PDFs",
                "regenerateStored" => true,
            })
        end

        if type == "native"
            # Populate natives
            exporter.addProduct("native",{
                "naming" => "guid",
                "path" => "Natives",
                "regenerateStored" => true,
            })
        end
    end

    items = current_case.search(search)
    logger.debug("Found #{items.length} items from search: #{search}")

    # Used to synchronize thread access in batch exported callback
	semaphore = Mutex.new

    # Setup batch exporter callback
    exporter.whenItemEventOccurs do |info|
        potential_failure = info.getFailure
        if !potential_failure.nil?
            event_item = info.getItem
            logger.error("Export failure for item: #{event_item.getGuid} : #{event_item.getLocalisedName}")
        end
        # Make the progress reporting have some thread safety
        semaphore.synchronize {
            logger.debug("Exporting item: #{info.get_stage}")
        }
    end

    exporter.export_items(items)

    FileUtils.rm_rf(dir)
    
    logger.info("Populate stores finished")
    logger.info("FINISHED")
end