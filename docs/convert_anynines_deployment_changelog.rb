#!/usr/bin/env ruby
require 'pp'
require 'fileutils'

# NOTE: cleanup old versions

def clean 
  FileUtils.rm(Dir.glob('changelog/*.md'))
end


# NOTE: create new entries
def convert
  metadata  = File.read('changelog/_metadata.yml')
  changelog = File.read('../anynines-deployment/CHANGELOG.md')
  newest    = nil


  changelog.split(/^\#\# /).each do |segment|
    match_data = segment.match(/^\[(.*)\] - (.*)/)

    if match_data
      version  = match_data[1]
      date     = match_data[2]
      newest ||= version

      puts "Found version #{version}"
      sanitize_content(content: segment, version: version, newest: newest)
      content  = <<~TEXT
#{metadata}
slug: changelog-#{version}
title: #{version}
date: #{date}
---

## #{segment}
      TEXT

      File.write("changelog/#{date}-version-#{version}.md", content, mode: 'w')
      puts "Generated version #{version}"
    end
  end

end

def sanitize_content(content: , version: , newest:)
  # remove double title
  content.gsub!(/^\[(.*)\] - (.*)/, '')

  # remove file links to github
  content.gsub!(
    /\[(?<title>.*)\]\((?!docs\/).*\)/i, 
    '\k<title>'
  )
  
  # TODO we will remove in the MVP all links

  content.gsub!(
    /\[(?<title>.*)\]\(.*\)/i, 
    '\k<title>'
  )


  # correct relative doc links
  #content.gsub!(
  #  /\[(?<title>.*)\]\((?<link>.*)\.md(?<anchor>.*)?\)/i, 
  #  '[\k<title>](\k<link>\k<anchor>)'
  #)

  # rebind links to version of the docs
  #if version != newest
  #  content.gsub!(
  #    /\[(?<title>.*)\]\(docs\/(?<link>.*)\)/i, 
  #    "[\\k<title>](/docs/#{version}/\\k<link>)"
  #  )
  #end
end


convert

