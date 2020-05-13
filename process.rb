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

# Obtain the case factory
case_factory = $utilities.getCaseFactory

# Create variable for case directory
case_directory = settings["case_directory"] + "\\" + settings["case_settings"]["name"] + Time.now.strftime("%Y%m%d%H%M%S")

# Create the case
nuix_case = case_factory.create(case_directory, settings["case_settings"])

# Obtain a processor
processor = nuix_case.create_processor

evidence_container = processor.new_evidence_container(settings["evidence_settings"]["name"])
evidence_container.add_file(settings["evidence_settings"]["directory"])
evidence_container.set_description(settings["evidence_settings"]["description"])
evidence_container.set_encoding(settings["evidence_settings"]["encoding"])
evidence_container.set_time_zone(settings["evidence_settings"]["time_zone"])
evidence_container.set_initial_custodian(settings["evidence_settings"]["custodian"])
evidence_container.set_locale(settings["evidence_settings"]["locale"])

evidence_container.save

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
exit(true)