require 'json'

module Minikube

  def self.does_demo_cluster_exist?
    json = JSON.parse(`minikube profile list -o json`)

    found = !json["valid"].select{ |cluster| cluster["Name"] == "a8s-demo" }.empty?

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
