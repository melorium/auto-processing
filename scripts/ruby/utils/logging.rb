#!/usr/bin/env ruby
require 'logger'
require 'json'


module Logging
  def self.get_logger(program)
    return Logger.new(
      STDOUT,
      level: Logger::DEBUG,
      progname: program,
      datetime_format: '%Y-%m-%dT%H:%M:%SZ',
      formatter: proc do |severity, datetime, progname, msg, exception|
        JSON.dump(level: severity.downcase, msg: msg, source: progname, time: "#{datetime.to_s}")+"\n"
      end
    )
  end
end
