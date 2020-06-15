require File.join(__dir__, "../utils", "logging")


$logger = Logging.get_logger("main.rb")

def main(settings, single_case, review_compound)
  for sub_step in settings
    if sub_step["type"] == "script"
    
      if sub_step["name"] == "search_and_tag"
        require File.join(__dir__, sub_step["name"])
        search_and_tag(single_case, sub_step["search"], sub_step["tag"], sub_step["files"])
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
        require File.join(__dir__, sub_step["name"])
        exclude(single_case, sub_step["search"], sub_step["reason"])
      end

      if sub_step["name"] == "reload"
        unless $utilities.get_processing_profile_store.contains_profile(sub_step["profile"])
          # Import the profile
          $logger.debug("Did not find the requested processing-profile in the profile-store")
          $logger.info("Importing new processing-profile #{sub_step["profile"]}}")
          $utilities.get_processing_profile_store.import_profile(sub_step["profile_location"], sub_step["profile"])
          $logger.info("Processing-profile has been imported")
        end
      
        # Set the profile to the processor
        reload_processor = single_case.create_processor
        reload_processor.set_processing_profile(sub_step["profile"])
        $logger.info("Processing-profile: #{sub_step["profile"]} has been set to the processor")
        require File.join(__dir__, sub_step["name"])
        reload(reload_processor, single_case, sub_step["search"])
      end
  
    end
  
    if sub_step["type"] == "sub_case"
      $logger.info("Creating subcase with search: " + sub_step["search"])
      sub_case = $utilities.get_case_subset_exporter
      items = single_case.search(sub_step["search"])
  
      $logger.info("Exporting subcase to: #{sub_step["case"]["directory"]}")
      sub_case.export_items(items, sub_step["case"]["directory"], sub_step["case"])
      $logger.info("Subcase exported")
  
      unless review_compound.nil?
        # Open the sub case as a Nuix Case
        sub_case_nuix = $utilities.getCaseFactory.open(sub_step["case"]["directory"])
        $logger.debug("Adding case: #{sub_step["case"]["name"]} to review-compound")

        # Add the sub case to the review compound
        review_compound.add_child_case(sub_case_nuix)
        $logger.info("Added case: #{sub_step["case"]["name"]} to review-compound")
      else
        $logger.debug("Did not add case: #{sub_step["case"]["name"]} to review-compound")
      end
    end
  end
end