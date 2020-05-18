require "time"
require "json"
require "optparse"


options = {}
OptionParser.new do |opts|
  opts.banner = "Usage: process.rb [options]"
  opts.on("-s", "--settings Path for process settings", "Settings") { |v| options[:settings] = v }
  opts.on("-p", "--profile Path for the processing-profile", "Processing Profile") { |v| options[:profile] = v }
  opts.on("-n", "--name for the processing-profile", "Profile Name") { |v| options[:profile_name] = v }

end.parse!

time_start = Time.now

file = File.read(options[:settings])
settings = JSON.parse(file)
case_settings = settings["case_settings"]
evidence_settings = settings["evidence_settings"]

# Obtain the case factory
case_factory = $utilities.getCaseFactory

# Check if the user wants to add the case to a compound
if settings["compound"]
  compound_settings = settings["compound_case"].merge!(compound: true)
  begin
    # Try to open the compound case
    compound = case_factory.open(compound_settings["directory"])
    
    # Handle the exception if the case doesnt exists
    rescue java.io.IOException
      puts("#{Time.now} Compound case doesnt exists - trying to create a new compound-case")
      
      begin
        # Try to create the new compound-case
        compound = case_factory.create(compound_settings["directory"], compound_settings)

        # Handle the exception and exit program
        rescue java.nio.file.FileAlreadyExistsException
          puts("#{Time.now} Cant create compound case - Directory already exists - choose a new compound directory")
          exit(false)
      end

  end
  puts("#{Time.now} Compound case opened")
end

# Create variable for the simple-case directory
case_directory = case_settings["directory"] + "\\" + case_settings["name"] + Time.now.strftime("%Y%m%d%H%M%S")

# Create the case
nuix_case = case_factory.create(case_directory, case_settings)

# Obtain a processor
processor = nuix_case.create_processor

# Loop through the different evidence and add it to their own container
for evidence in evidence_settings
  evidence_container = processor.new_evidence_container(evidence["name"])
  evidence_container.add_file(evidence["directory"])
  evidence_container.set_description(evidence["description"])
  evidence_container.set_encoding(evidence["encoding"])
  evidence_container.set_time_zone(evidence["time_zone"])
  evidence_container.set_initial_custodian(evidence["custodian"])
  evidence_container.set_locale(evidence["locale"])

  evidence_container.save
end

# Check if the profile exists in the store
unless $utilities.get_processing_profile_store.contains_profile(options[:profile_name])
    # Import the profile
    puts("#{Time.now} Did not find the requested processing-profile in the profile-store")
    puts("#{Time.now} Importing new processing-profile #{options[:profile_name]}}")
    $utilities.get_processing_profile_store.import_profile(options[:profile], options[:profile_name])
    puts("#{Time.now} Processing-profile has been imported")
end
# Set the profile to the processor
processor.set_processing_profile(options[:profile_name])
puts("#{Time.now} Processing-profile: #{options[:profile_name]} has been set to the processor")


# Count how many items have been processed
processed_item_count = 0
processor.when_item_processed do |processed_item|
	processed_item_count += 1
	puts ("#{Time.now} Item processed: #{processed_item.get_mime_type()} Count: #{processed_item_count}")
end

puts ("#{Time.now} Processing started...")
processor.process
puts ("#{Time.now} Processing has completed")

time_end = Time.now

execution_time = time_end - time_start

puts("#{Time.now} Execution time: " + execution_time.to_i.to_s + " seconds")


# check if the case is compound just to be sure
if compound.is_compound()
  compound.add_child_case(nuix_case) # Add the newly processed case to the compound-case
  puts("#{Time.now} Added case to compound")
  compound.close()
else
  puts("#{Time.now} Did not add this case to a compound")
end

exit(true)