#!/usr/bin/env ruby
require 'pp'
require 'json'


def build_json
 versions =  Dir.glob('versioned_docs/*').map do |version_folder|
    version_folder.match(/version\-(.*)\z/)[1]
 end.sort.reverse[0..4]

  File.write("versions.json", JSON.generate(versions))
end

build_json
