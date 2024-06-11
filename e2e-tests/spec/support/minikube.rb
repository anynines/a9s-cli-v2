require 'json'

module Minikube

  def self.create_cluster(cluster_name = 'a9s-create-stack-rspec')
    output = `minikube start -p #{cluster_name} --nodes 3 --memory 4gb -o json`

    output.each_line do |line|
      json = JSON.parse(line)

      return true if json["data"] && json["data"]["name"] && json["data"]["name"] == "Done"
    end

    puts "Something went wrong:\n#{output}"

    return false
  end

  def self.delete_cluster(cluster_name = 'a9s-create-stack-rspec')
    output = `minikube delete -p #{cluster_name} -o json`

    if $?.success?
      return true
    end

    puts "Something went wrong:\n#{output}"
    return false
  end

  def self.does_demo_cluster_exist?
    does_cluster_exist?("a8s-demo")
  end

  def self.does_cluster_exist?(cluster_name)
    json = JSON.parse(`minikube profile list -o json`)

    found = !json["valid"].select{ |cluster| cluster["Name"] == cluster_name }.empty?

    return found
  end

  def self.delete_demo_cluster
    puts "Deleting a8s-demo minikube cluster..."
    ret = system("minikube delete --profile a8s-demo")

    unless ret then
      raise "Couldn't delete minikube a8s-demo cluster"
    end
    puts "Done"
  end
end
