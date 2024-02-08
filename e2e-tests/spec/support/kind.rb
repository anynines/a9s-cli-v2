module Kind
  def self.does_demo_cluster_exist?
    `kind get clusters`.split("\n").include?("a8s-demo")
  end

  def self.delete_demo_cluster
    unless system("kind delete cluster -n a8s-demo") then
      raise "Couldn't delete kind cluster a8s-demo"
    end
  end
end
