require "time"
require "json"
require "optparse"
require File.join(__dir__, "utils", "logging") 


class Case
    def initialize(settings)
        @case_factory = $utilities.getCaseFactory
        @settings = settings
    end

    def create_compound(compound)
        compound_settings = compound.merge!(compound: true)
        
        begin
            # Try to open the compound case
            compound_case = @case_factory.open(compound_settings["directory"])
        
        # Handle the exception if the case doesnt exists
        rescue java.io.IOException
            $logger.warn("Compound case doesnt exists - trying to create a new compound-case")
            
            begin
                # Try to create the new compound-case
                compound_case = @case_factory.create(compound_settings["directory"], compound_settings)

            rescue java.io.IOException => exception
                $logger.fatal("problem creating new case, case might already be open: #{exception.backtrace}")
                exit(false)

            # Handle the exception and exit program
            rescue java.nio.file.FileAlreadyExistsException
                $logger.fatal("Cant create compound case - Directory already exists - choose a new compound directory")
                exit(false)
            end

            $logger.info("Compound case opened")
        end
        return compound_case
    end

    def create_single
        # Create variable for the simple-case directory
        case_name = @settings["case_settings"]["name"]
        case_directory = @settings["case_settings"]["directory"]
        $logger.info("creating single-case to: #{case_directory}")
        return @case_factory.create(case_directory, @settings["case_settings"])
    end

    def tear_down(compound_case, single_case, review_compound)
        # check if the case is compound just to be sure
        unless compound_case.nil?
          $logger.debug("Adding case to compound")
          compound_case.add_child_case(single_case) # Add the newly processed case to the compound-case
          $logger.info("Added case to compound")
          compound_case.close()
          review_compound.close()
        else
          $logger.debug("Did not add case to compound")
        end
    end
end

class Processor
    def initialize(settings, single_case)
        @settings = settings
        @processor = single_case.create_processor
    end

    def add_evidence
        $logger.info("Adding evidence to container")
        if @settings["evidence_settings"].nil?
            $logger.fatal("No evidence is set to this config")
            exit(false)
        end
        for evidence in @settings["evidence_settings"]
            evidence_container = @processor.new_evidence_container(evidence["name"])
            evidence_container.add_file(evidence["directory"])
            evidence_container.set_description(evidence["description"])
            evidence_container.set_encoding(evidence["encoding"])
            evidence_container.set_time_zone(evidence["time_zone"])
            evidence_container.set_initial_custodian(evidence["custodian"])
            evidence_container.set_locale(evidence["locale"])
            
            evidence_container.save
            $logger.debug("Added evidence: #{evidence["name"]}")
        end
        $logger.debug("Total of #{@settings["evidence_settings"].length} evidence added to container")

    end

    def set_profile
        $logger.info("Puts the profile to the processor")
        # Set the profile to the processor
        @processor.set_processing_profile(@settings["profile"])
        $logger.info("Processing-profile: #{@settings["profile"]} has been set to the processor")
    end

    def run
        # Count how many items have been processed
        processed_item_count = 0
        @processor.when_item_processed do |processed_item|
          processed_item_count += 1
          $logger.debug("Item processed: #{processed_item.get_mime_type()} Count: #{processed_item_count}")
        end
      
        $logger.info("Processing started...")
        @processor.process
        $logger.info("Processing has completed")
    end
        
    def run_scripts(single_case, compound_case)
        unless @settings["sub_steps"].nil?
            require File.join(@settings["working_path"], "scripts\\ruby\\scripts", "main.rb")
            main(@settings["sub_steps"], single_case, compound_case, $current_cfg)
        end
    end
end

def get_settings
    options = {}
    OptionParser.new do |opts|
        opts.banner = "Usage: process.rb [options]"
        opts.on("-s", "--settings Path for process settings", "Settings") { |v| options[:settings] = v }
        opts.on("-c", "--config name of the current running config", "Config") { |v| options[:config] = v }
    end.parse!

    $current_cfg = options[:config]

    file = File.read(options[:settings])
    return JSON.parse(file)
end

begin
    time_start = Time.now 
    settings = get_settings

    $logger = Logging.get_logger("process.rb", $current_cfg)
    $logger.info("START")

    case_factory = Case.new(settings)
    if settings["compound"]
        compound_case = case_factory.create_compound(settings["compound_case"])
        review_compound = case_factory.create_compound(settings["review_compound"])
    end
    single_case = case_factory.create_single

    processor = Processor.new(settings, single_case)
    processor.add_evidence
    processor.set_profile
    processor.run
    processor.run_scripts(single_case, review_compound)

    case_factory.tear_down(compound_case, single_case, review_compound)

    time_end = Time.now
    execution_time = time_end - time_start
    $logger.info("Execution time: " + execution_time.to_i.to_s + " seconds")
    $logger.info("FINISHED")

rescue => e
    $logger.fatal("Failed to run script: #{e}")
    exit(false)
end

exit(true)