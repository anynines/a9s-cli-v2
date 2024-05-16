require_relative '../spec_helper'
require_relative '../support/minikube'
require_relative '../support/kind'

RSpec.shared_context "a8s-pg", :shared_context => :metadata, order: :defined do
  before(:context) do
    @workload_namespace = "a8s-workload"
    @service_instance_name = "clustered"
    @backup_name ||= "clustered-bu"
    @restore_name ||= "clustered-rs"
    @sql_file_small = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "assets", "pg_demo_data_small.sql"))
    @service_binding_name ||= "clustered-sb"
  end

  context "service instances" do
    it "creates a clustered a8s pg service instance" do
      cmd = "a9s create pg instance --name #{@service_instance_name} --replicas 3 -n #{@workload_namespace} --yes"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("The #{@service_instance_name} appears to be ready. All expected pods are running.")
    end
  end

  context "load SQL data" do
    it "loads SQL data into the a8s pg service instance" do
      cmd = "a9s pg apply --file #{@sql_file_small} --service-instance #{@service_instance_name} -n #{@workload_namespace} --yes"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Successfully applied SQL file to pod")
    end

    it "verifies that data has been loaded into the a8s pg service instance" do
      cmd = "a9s pg apply --service-instance #{@service_instance_name} -n #{@workload_namespace} --sql \"SELECT COUNT(*) FROM POSTS\" --yes"

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to match(/-------\n\s+20/)
    end
  end

  context "service bindings" do
    it "creates a service binding for the a8s pg service instance" do
      cmd = "a9s create pg servicebinding --name #{@service_binding_name} --service-instance #{@service_instance_name} -n #{@workload_namespace} --yes"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("The service binding has been created successfully.")
    end

    it "deletes a given service binding" do
      cmd = "a9s delete pg servicebinding --name #{@service_binding_name} -n #{@workload_namespace} --yes"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("The service binding has been deleted successfully.")

      #TODO Add a test that verifies whether the SB has actually been deleted
    end
  end

  context "backups" do
    it "creates a backup of an a8s pg service instance" do
      cmd = "a9s create pg backup --name #{@backup_name} -i #{@service_instance_name} -n #{@workload_namespace} --yes"

      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("The backup with the name #{@backup_name} in namespace #{@workload_namespace} has been successful")
    end

    it "fails to create a backup of a non-existing service instance" do


      cmd = "a9s create pg backup --name #{@backup_name} -i nonexistinginstance -n #{@workload_namespace} --yes"

      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Can't create backup for non-existing service instance")
    end
  end

  context "restore" do
    it "creates a restore of a backup" do
      cmd = "a9s create pg restore --name #{@restore_name} -b #{@backup_name} -i #{@service_instance_name} -n #{@workload_namespace} --yes"

      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("The restore with the name #{@restore_name} in namespace #{@workload_namespace} has been successful.")
    end

    it "fails to create a restore of a non existing service instance" do
      cmd = "a9s create pg restore --name #{@restore_name} -b nonobackup -i #{@service_instance_name} -n #{@workload_namespace} --yes"

      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Can't create restore for non-existing backup")
    end

    it "fails to create a restore of a non existing backup" do
      cmd = "a9s create pg restore --name #{@restore_name} -b #{@backup_name} -i idontexist -n #{@workload_namespace} --yes"

      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Can't create restore for non-existing service instance")
    end
  end
end


RSpec.describe "a9s-cli" do
  context "about the a9s executable" do
    it "verifies the existence of an a9s executable" do

      cmd = "a9s --help"
      logger.info(cmd)
      # surpresses stdout with File::NULL
      ret = system(cmd, :out => File::NULL)

      expect(ret).to be(true)
    end

    it "verifies the output of a9s --help to see if it is the right a9s command" do
      cmd = "a9s --help"
      logger.info(cmd)

      ret = `#{cmd}`

      logger.info "\t" + ret.to_s

      expect(ret).to include("A tool to make the use of a9s Platform modules more enjoyable.")
    end

    it "verifies the execution of a9s version" do
      cmd = "a9s version"
      logger.info(cmd)

      ret = `#{cmd}`

      logger.info "\t" + ret.to_s

      expect(ret).to include("a9s cli version:")
    end

    it "verifies the output of a9s version against the version specified in the VERSION file check for consistency between the spec and binary" do
      cmd = "a9s version"
      logger.info(cmd)

      ret = `#{cmd}`

      logger.info "\t" + ret.to_s

      expect(ret).to include(File.read(File.expand_path("../../../../VERSION", __FILE__)).chomp)
    end
  end

  # Idea: use contexts to immitate the a9s command topology
  context "create", order: :defined do
    context "stack", order: :defined do
      before :context do
        Minikube.create_cluster
      end

      it "creates an a8s stack on a given Kubernetes cluster", :clusterop => true, :slow => true do
            cmd = "a9s create stack a8s -c a9s-create-stack-rspec --yes"

            logger.info cmd

            output = `#{cmd}`

            logger.info "\t" + output

            expect(output).to include("You are now ready to create a8s Postgres service instances.")
            kubectl_create_namespace(@workload_namespace)
          end
          include_context "a8s-pg", :include_shared => true

      after :context do
        Minikube.delete_cluster
      end
    end
    context "cluster", order: :defined do
      context "a8s", order: :defined do
        context "kind", order: :defined, kind: true do
          before (:context) do
            @backup_name = "kind-clustered-bu"
            if Kind::does_demo_cluster_exist? then
              logger.info "Deleting existing kind cluster..."
              Kind::delete_demo_cluster
            end
          end
          it "creates a Kubernetes cluster with an a8s stack", :clusterop => true, :slow => true do
            cmd = "a9s create cluster a8s -p kind --yes"

            logger.info cmd

            output = `#{cmd}`

            logger.info "\t" + output

            expect(output).to include("You are now ready to create a8s Postgres service instances.")
            kubectl_create_namespace(@workload_namespace)
          end
          include_context "a8s-pg", :include_shared => true

          context "delete cluster", :order => :defined do
              it "delete the a8s cluster cluster", :clusterop => true do
                cmd = "a9s delete cluster a8s -p kind --yes"
                logger.info cmd

                ret = system(cmd)

                expect(ret).to be(true)
                expect(Kind::does_demo_cluster_exist?).to be(false)
              end
            end
        end
        context "minikube", order: :defined, minikube: :true do
          context "use cluster" do
            before(:context) do
              @backup_name = "minikube-clustered-bu"
              if Minikube::does_demo_cluster_exist? then
                logger.info "Found a minikube demo cluster."
                Minikube::delete_demo_cluster
              end
            end

            it "creates an a8s cluster cluster", :clusterop => true, :slow => true do
              cmd = "a9s create cluster a8s -p minikube --yes"

              logger.info cmd
              output = `#{cmd}`
              logger.info "\t" + output

              expect(output).to include("You are now ready to create a8s Postgres service instances.")
              kubectl_create_namespace(@workload_namespace)
            end

            include_context "a8s-pg", :include_shared => true

            context "delete cluster", :order => :defined do
              it "delete the a8s cluster cluster", :clusterop => true do
                cmd = "a9s delete cluster a8s -p minikube --yes"

                logger.info cmd

                ret = system(cmd)

                logger.info "\t" + ret.to_s

                expect(ret).to be(true)
                expect(Minikube::does_demo_cluster_exist?).to be(false)
              end
            end
          end
        end
      end
    end
  end
end
