require File.join(__dir__, "../utils", "logging")


$logger = Logging.get_logger("main.rb")

def main(settings, single_case)
  for sub_step in settings
    if sub_step["type"] == "script"
    
      if sub_step["name"] == "search_and_tag"
        require File.join(__dir__, sub_step["name"])
        search_and_tag(single_case, sub_step["search"], sub_step["tag"], "C:\\Users\\sja\\Download\\Attachment.json")
      end
  
      if sub_step["name"] == "ocr"
        # Check if the profile exists in the store
        unless $utilities.get_ocr_profile_store.contains_profile(sub_step["profile"])
          # Import the profile
          $logger.debug("Did not find the requested ocr-profile in the profile-store")
          $logger.info("Importing new ocr-profile #{sub_step["profile"]}}")
          $utilities.get_ocr_profile_store.import_profile(sub_step["profile_location"], sub_step["profile"])
          $logger.info("OCR-profile has been imported")
        end
        # Set the profile to the processor
        ocr_processor = $utilities.createOcrProcessor
        ocr_processor.set_ocr_profile(sub_step["profile"])
        $logger.info("OCR-profile: #{sub_step["profile"]} has been set to the processor")
  
        require File.join(__dir__, sub_step["name"])
        ocr(ocr_processor, single_case, sub_step["search"])
      end
  
      if sub_step["name"] == "exclude"
        $logger.error("DIR:"+ __dir__)
        require File.join(__dir__, sub_step["name"])
        exclude(single_case, sub_step["search"], sub_step["reason"])
      end
  
    end
  
    if sub_step["type"] == "sub_case"
      
      sub_directory = sub_step["export_path"] + "\\" + single_case.get_name + " - REVIEW"
      if sub_step["export_path"].length < 1
        $logger.warn("Export path not provided for subcase export")
        sub_directory = case_directory + " - REVIEW"
      end
      
      $logger.info("Creating subcase with search: " + sub_step["search"])
      sub_case = $utilities.get_case_subset_exporter
      items = single_case.search(sub_step["search"])
  
      $logger.info("Exporting subcase to: #{sub_directory}")
      sub_case.export_items(items, sub_directory)
      $logger.info("Subcase exported")
    end
  end
end