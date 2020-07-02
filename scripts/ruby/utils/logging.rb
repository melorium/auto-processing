#!/usr/bin/env ruby
require 'logger'
require 'json'


module Logging
  def self.get_logger(program, cfg)
    datetime_format = "%FT%T.%3N#{(Time.new.utc? ? 'Z' : '%:z')}"
    return Logger.new(
      STDOUT,
      level: Logger::DEBUG,
      progname: program,
      formatter: proc do |severity, datetime, progname, msg, exception|
      JSON.dump(level: severity.downcase, msg: msg, source: progname, config: cfg, time: "#{datetime.strftime(datetime_format)}")+"\n"
    end
    )
  end
end
