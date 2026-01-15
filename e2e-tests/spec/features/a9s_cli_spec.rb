require_relative '../spec_helper'
require_relative '../support/minikube'
require_relative '../support/kind'

require 'fileutils'
require 'tmpdir'

RSpec.shared_context "a8s-pg", :shared_context => :metadata, order: :defined do
  before(:context) do

    # Authenticate at the beiginning in case of 1password is being used
    @workload_namespace = "a8s-workload"
    @service_instance_name = "clustered"
    @backup_name ||= "clustered-bu"
    @restore_name ||= "clustered-rs"
    @sql_file_small = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "assets", "pg_demo_data_small.sql"))
    @service_binding_name ||= "clustered-sb"
  end

  context "service instances" do
    it "creates a clustered a8s pg service instance" do
      sleep(10)
      cmd = "a9s create pg instance --name #{@service_instance_name} --replicas 3 -n #{@workload_namespace} --verbose --yes"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("The #{@service_instance_name} system appears to be ready. All expected pods are running.")
    end
  end

  context "load SQL data" do
    it "loads SQL data into the a8s pg service instance" do
      cmd = "a9s pg apply --file #{@sql_file_small} --service-instance #{@service_instance_name} -n #{@workload_namespace} --verbose --yes"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Successfully applied SQL file to pod")
    end

    it "verifies that data has been loaded into the a8s pg service instance" do
      cmd = "a9s pg apply --service-instance #{@service_instance_name} -n #{@workload_namespace} --sql \"SELECT COUNT(*) FROM POSTS\" --yes"

      output = `#{cmd}`

      logger.info "\t" + output

      safe_output = output.encode("UTF-8", invalid: :replace, undef: :replace, replace: "")
      expect(safe_output).to match(/-------\n\s+20/)
    end
  end

  context "service bindings" do
    it "creates a service binding for the a8s pg service instance" do
      cmd = "a9s create pg servicebinding --name #{@service_binding_name} --service-instance #{@service_instance_name} -n #{@workload_namespace} --verbose --yes"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      safe_output = output.encode("UTF-8", invalid: :replace, undef: :replace, replace: "")
      expect(safe_output).to include("The service binding has been created successfully.")
      
      # Verify the service binding actually exists in Kubernetes
      kubectl_verify_service_binding_exists(@service_binding_name, @workload_namespace)
    end

    it "deletes a given service binding" do
      cmd = "a9s delete pg servicebinding --name #{@service_binding_name} -n #{@workload_namespace} --verbose --yes"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      safe_output = output.encode("UTF-8", invalid: :replace, undef: :replace, replace: "")
      expect(safe_output).to include("The service binding has been deleted successfully.")

      # Verify the service binding has actually been deleted from Kubernetes
      kubectl_verify_service_binding_not_exists(@service_binding_name, @workload_namespace)
    end
  end

  context "backups" do
    it "creates a backup of an a8s pg service instance" do
      cmd = "a9s create pg backup --name #{@backup_name} -i #{@service_instance_name} -n #{@workload_namespace} --verbose --yes"

      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("The backup with the name #{@backup_name} in namespace #{@workload_namespace} has been successful")
    end

    it "fails to create a backup of a non-existing service instance" do

      cmd = "a9s create pg backup --name #{@backup_name} -i nonexistinginstance -n #{@workload_namespace} --verbose --yes"

      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Can't create backup for non-existing service instance")
    end
  end

  context "restore" do
    it "creates a restore of a backup" do
      cmd = "a9s create pg restore --name #{@restore_name} -b #{@backup_name} -i #{@service_instance_name} -n #{@workload_namespace} --verbose --yes"

      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("The restore with the name #{@restore_name} in namespace #{@workload_namespace} has been successful.")
    end

    it "fails to create a restore of a non existing service instance" do
      cmd = "a9s create pg restore --name #{@restore_name} -b nonobackup -i #{@service_instance_name} -n #{@workload_namespace} --verbose --yes"

      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Can't create restore for non-existing backup")
    end

    it "fails to create a restore of a non existing backup" do
      cmd = "a9s create pg restore --name #{@restore_name} -b #{@backup_name} -i idontexist -n #{@workload_namespace} --verbose --yes"

      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Can't create restore for non-existing service instance")
    end
  end

  context "service instances cleanup" do
    it "deletes the a8s pg service instance" do
      cmd = "a9s delete pg instance --name #{@service_instance_name} -n #{@workload_namespace} --verbose --yes"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Service instance #{@service_instance_name} successfully deleted from namespace #{@workload_namespace}.")
      kubectl_verify_pg_service_instance_not_exists(@service_instance_name, @workload_namespace)
    end

    it "warns when deleting a non-existing a8s pg service instance" do
      cmd = "a9s delete pg instance --name #{@service_instance_name} -n #{@workload_namespace} --verbose --yes"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Can't delete service instance. Service instance #{@service_instance_name} doesn't exist in namespace #{@workload_namespace}!")
    end

    it "fails to create a service binding for a non-existing service instance" do
      cmd = "a9s create pg servicebinding --name #{@service_binding_name} --service-instance #{@service_instance_name} -n #{@workload_namespace} --verbose --yes"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      safe_output = output.encode("UTF-8", invalid: :replace, undef: :replace, replace: "")
      expect(safe_output).to include("Can't create service binding for non-existing service instance #{@service_instance_name} in namespace #{@workload_namespace}")
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

  context "cluster pwd" do
    it "prints the configured working directory from config" do
      temp_home = Dir.mktmpdir("a9s-home")
      workdir = File.join(temp_home, "workdir")
      config_path = File.join(temp_home, ".a9s")
      File.write(config_path, "WorkingDir: #{workdir}\nDemoSpace: a8s-demo\n")

      cmd = "HOME=#{temp_home} a9s cluster pwd"
      logger.info(cmd)
      ret = `#{cmd}`
      logger.info "\t" + ret.to_s

      expect(ret).to eq(workdir)
    ensure
      FileUtils.remove_entry(temp_home) if temp_home && File.exist?(temp_home)
    end
  end

  context "pg apply" do
    it "fails when neither --file nor --sql is provided" do
      cmd = "a9s pg apply --service-instance missing -n default --yes"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please supply either --sql with an SQL statement or --file with a path to a sql file.")
    end
  end

  context "use" do
    it "requires cluster-name for klutch" do
      cmd = "a9s use klutch"
      logger.info(cmd)

      output = `#{cmd} 2>&1`

      logger.info "\t" + output

      expect(output).to include("--cluster-name is required")
    end

    it "prompts for a subcommand" do
      cmd = "a9s use"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select a subcommand from the list below.")
    end
  end

  context "estimate-cost" do
    it "requires provider for klutch cost estimate" do
      cmd = "a9s estimate-cost cluster klutch"
      logger.info(cmd)

      output = `#{cmd} 2>&1`

      logger.info "\t" + output

      expect(output).to include("Please select a provider via -p")
    end

    it "prompts for a resource type" do
      cmd = "a9s estimate-cost"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please choose a resource to estimate.")
    end
  end

  context "get" do
    it "prompts for a subcommand" do
      cmd = "a9s get"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select a subcommand from the list below.")
    end
  end

  context "get clusters" do
    it "prompts for a subcommand" do
      cmd = "a9s get clusters"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select a subcommand from the list below.")
    end
  end

  context "get klutch" do
    it "prompts for a subcommand" do
      cmd = "a9s get klutch"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select a subcommand from the list below.")
    end
  end

  context "bind" do
    it "prompts for a subcommand" do
      cmd = "a9s bind"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select a subcommand from the list below.")
    end
  end

  context "bind klutch workload" do
    it "requires control-plane URL" do
      cmd = "a9s bind klutch workload"
      logger.info(cmd)

      output = `#{cmd} 2>&1`

      logger.info "\t" + output

      expect(output).to include("control-plane URL is required")
    end
  end

  context "bind klutch" do
    it "prompts for a subcommand" do
      cmd = "a9s bind klutch"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select a subcommand from the list below.")
    end
  end

  context "klutch" do
    it "prompts for a klutch subcommand" do
      cmd = "a9s klutch"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select a subcommand from the list below.")
    end
  end

  context "apply" do
    it "prompts for a subcommand" do
      cmd = "a9s apply"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select a subcommand from the list below.")
    end
  end

  context "apply klutch" do
    it "prompts for a subcommand" do
      cmd = "a9s apply klutch"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select a subcommand from the list below.")
    end
  end

  context "apply klutch control-plane" do
    it "requires hosted zone name" do
      cmd = "a9s apply klutch control-plane"
      logger.info(cmd)

      output = `#{cmd} 2>&1`

      logger.info "\t" + output

      expect(output).to include("The --hosted-zone-name flag is required until self-signed certificates are supported.")
    end
  end

  context "delete" do
    it "prompts for a resource type" do
      cmd = "a9s delete"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select the data service resource type you would like to delete.")
    end
  end

  context "pg" do
    it "prompts for a subcommand" do
      cmd = "a9s pg"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select a subcommand from the list below.")
    end
  end

  context "create" do
    it "prompts for a resource type" do
      cmd = "a9s create"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select the data service resource type you would like to instantiate.")
    end
  end

  context "cluster" do
    it "prompts for a cluster sub-command" do
      cmd = "a9s cluster"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select a cluster sub-command.")
    end
  end

  context "create pg" do
    it "prompts for a PostgreSQL resource" do
      cmd = "a9s create pg"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select a PostgreSQL resource such as (service) instance.")
    end
  end

  context "delete pg" do
    it "prompts for a PostgreSQL resource" do
      cmd = "a9s delete pg"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select a PostgreSQL resource such as (service) instance.")
    end
  end

  context "create cluster" do
    it "prompts for a sub-command" do
      cmd = "a9s create cluster"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please use a sub-command.")
    end
  end

  context "create cluster klutch" do
    it "prompts for a sub-command" do
      cmd = "a9s create cluster klutch"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please use a sub-command.")
    end
  end

  context "create cluster klutch control-plane" do
    it "requires a provider" do
      cmd = "a9s create cluster klutch control-plane"
      logger.info(cmd)

      output = `#{cmd} 2>&1`

      logger.info "\t" + output

      expect(output).to include("Please select a provider via -p.")
    end
  end

  context "create cluster klutch workload" do
    it "requires a provider" do
      cmd = "a9s create cluster klutch workload"
      logger.info(cmd)

      output = `#{cmd} 2>&1`

      logger.info "\t" + output

      expect(output).to include("Please select a provider via -p.")
    end
  end

  context "create stack" do
    it "prompts for a sub-command" do
      cmd = "a9s create stack"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please use a sub-command.")
    end
  end

  context "create klutch" do
    it "prompts for a sub-command" do
      cmd = "a9s create klutch"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please use a sub-command.")
    end
  end

  context "create klutch tenant" do
    it "requires a tenant name flag" do
      cmd = "a9s create klutch tenant"
      logger.info(cmd)

      output = `#{cmd} 2>&1`

      logger.info "\t" + output

      expect(output).to include("The --tenant-name flag is required.")
    end
  end

  context "delete cluster" do
    it "prompts for a sub-command" do
      cmd = "a9s delete cluster"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Use a sub-command to choose the demo resource to be deleted.")
    end
  end

  context "delete cluster klutch" do
    it "prompts for control-plane or workload" do
      cmd = "a9s delete cluster klutch"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select either the control-plane or workload subcommand.")
    end
  end

  context "delete klutch" do
    it "prompts for a subcommand" do
      cmd = "a9s delete klutch"
      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      expect(output).to include("Please select a subcommand from the list below.")
    end
  end

  context "get klutch tenant" do
    it "requires a tenant name argument" do
      cmd = "a9s get klutch tenant"
      logger.info(cmd)

      output = `#{cmd} 2>&1`

      logger.info "\t" + output

      expect(output).to include("accepts 1 arg(s)")
    end
  end

  context "delete tenant" do
    it "requires a tenant name argument" do
      cmd = "a9s delete tenant"
      logger.info(cmd)

      output = `#{cmd} 2>&1`

      logger.info "\t" + output

      safe_output = output.encode("UTF-8", invalid: :replace, undef: :replace, replace: "")
      expect(safe_output).to include("accepts 1 arg(s)")
    end
  end

  # Idea: use contexts to immitate the a9s command topology
  context "create", order: :defined do
    context "stack", order: :defined do
      before :context do
        @minikube_stack_cluster_name = "a9s-create-stack-rspec"

        if reuse_cluster?
          unless Minikube::does_cluster_exist?(@minikube_stack_cluster_name)
            skip("Skipping stack tests. Reuse requested but Minikube cluster #{@minikube_stack_cluster_name} not found.")
          end
          kubectl_ensure_namespace("a8s-workload")
        else
          logger.info "About to create Minikube cluster to enable stack testing..."

          if Minikube::does_cluster_exist?(@minikube_stack_cluster_name) then
            logger.info "Found a #{@minikube_stack_cluster_name} cluster. Deleting it..."
            Minikube::delete_cluster(@minikube_stack_cluster_name)
            logger.info "Done deleting #{@minikube_stack_cluster_name} cluster."
          end

          delete_a9s_backup_store_config

          logger.info "Creating Minikube cluster to enable stack testing..."
          unless Minikube.create_cluster then
            raise "Couldn't create Minikube cluster."
          end

          logger.info "Minikube cluster created."
        end
      end


      it "creates an a8s stack on a given Kubernetes cluster", :clusterop => true, :slow => true do
            skip("Skipping stack creation because A9S_E2E_REUSE_CLUSTER=1") if reuse_cluster?
            logger.info "Creating stack a8s..."

            cmd = "a9s create stack a8s -c #{@minikube_stack_cluster_name} --verbose --yes"

            logger.info cmd

            output = `#{cmd}`

            logger.info "Done creating stack:"

            logger.info "\t" + output

            expect(output).to include("You are now ready to create a8s Postgres service instances.")
            kubectl_create_namespace(@workload_namespace)
          end
          include_context "a8s-pg", :include_shared => true

      after :context do
        Minikube.delete_cluster(@minikube_stack_cluster_name) unless reuse_cluster?
      end
    end
    context "cluster", order: :defined do
      context "a8s", order: :defined do
        context "kind", order: :defined, kind: true do

          before (:context) do
            @backup_name = "kind-clustered-bu"
            if reuse_cluster?
              unless Kind::does_demo_cluster_exist?
                skip("Skipping kind tests. Reuse requested but kind cluster a8s-demo not found.")
              end
              kubectl_ensure_namespace("a8s-workload")
            else
              if Kind::does_demo_cluster_exist? then
                logger.info "Deleting existing kind cluster..."
                Kind::delete_demo_cluster
              end

              delete_a9s_backup_store_config
            end
          end

          it "creates a Kubernetes cluster with an a8s stack", :clusterop => true, :slow => true do
            skip("Skipping cluster creation because A9S_E2E_REUSE_CLUSTER=1") if reuse_cluster?
            cmd = "a9s create cluster a8s -p kind --verbose --yes"

            logger.info cmd

            output = `#{cmd}`

            logger.info "\t" + output

            expect(output).to include("You are now ready to create a8s Postgres service instances.")
            kubectl_create_namespace(@workload_namespace)
          end

          include_context "a8s-pg", :include_shared => true

          context "delete cluster", :order => :defined do
              it "delete the a8s cluster cluster", :clusterop => true do
                skip("Skipping cluster deletion because A9S_E2E_REUSE_CLUSTER=1") if reuse_cluster?
                cmd = "a9s delete cluster a8s -p kind --verbose --yes"
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
              if reuse_cluster?
                unless Minikube::does_demo_cluster_exist?
                  skip("Skipping minikube tests. Reuse requested but minikube cluster a8s-demo not found.")
                end
                kubectl_ensure_namespace("a8s-workload")
              else
                if Minikube::does_demo_cluster_exist? then
                  logger.info "Found a minikube demo cluster."
                  Minikube::delete_demo_cluster
                end

                delete_a9s_backup_store_config
              end
            end

            it "creates an a8s cluster cluster", :clusterop => true, :slow => true do
              skip("Skipping cluster creation because A9S_E2E_REUSE_CLUSTER=1") if reuse_cluster?
              cmd = "a9s create cluster a8s -p minikube --verbose --yes"

              logger.info cmd
              output = `#{cmd}`
              logger.info "\t" + output

              expect(output).to include("You are now ready to create a8s Postgres service instances.")
              kubectl_create_namespace(@workload_namespace)
            end

            include_context "a8s-pg", :include_shared => true

            context "delete cluster", :order => :defined do
              it "delete the a8s cluster cluster", :clusterop => true do
                skip("Skipping cluster deletion because A9S_E2E_REUSE_CLUSTER=1") if reuse_cluster?
                cmd = "a9s delete cluster a8s -p minikube --verbose --yes"

                logger.info cmd

                ret = system(cmd)

                logger.info "\t" + ret.to_s

                expect(ret).to be(true)
                expect(Minikube::does_demo_cluster_exist?).to be(false)
              end
            end
          end
        end
        context "minikube with AWS S3", order: :defined, minikube: :true do
          context "use cluster" do
            before(:context) do
              @aws_s3_bucket = ENV["AWS_S3_BUCKET_NAME"]
              @aws_access_key = ENV["AWS_ACCESSKEYID"]
              @aws_secret_key = ENV["AWS_SECRETKEY"]
              missing = []
              missing << "AWS_S3_BUCKET_NAME" if @aws_s3_bucket.to_s.empty?
              missing << "AWS_ACCESSKEYID" if @aws_access_key.to_s.empty?
              missing << "AWS_SECRETKEY" if @aws_secret_key.to_s.empty?
              skip("Skipping AWS S3 minikube e2e tests. Missing env vars: #{missing.join(', ')}") unless missing.empty?

              @backup_name = "minikube-clustered-bu"
              if reuse_cluster?
                unless Minikube::does_demo_cluster_exist?
                  skip("Skipping minikube AWS S3 tests. Reuse requested but minikube cluster a8s-demo not found.")
                end
                kubectl_ensure_namespace("a8s-workload")
              else
                if Minikube::does_demo_cluster_exist? then
                  logger.info "Found a minikube demo cluster."
                  Minikube::delete_demo_cluster
                end

                delete_a9s_backup_store_config
              end
            end

            it "creates an a8s cluster cluster", :clusterop => true, :slow => true do
              skip("Skipping cluster creation because A9S_E2E_REUSE_CLUSTER=1") if reuse_cluster?
              cmd = 'a9s create cluster a8s -p minikube --verbose --yes' \
                  ' --backup-provider="AWS"' \
                  ' --backup-region="eu-central-1"'
              cmd += ' --backup-bucket="'
              cmd += @aws_s3_bucket
              cmd += "\""
              cmd += ' --backup-store-accesskey="'
              cmd += @aws_access_key
              cmd += "\""
              cmd += ' --backup-store-secretkey="'
              cmd += @aws_secret_key
              cmd += "\""

              logger.info cmd
              output = `#{cmd}`
              logger.info "\t" + output

              expect(output).to include("You are now ready to create a8s Postgres service instances.")
              kubectl_create_namespace(@workload_namespace)
            end

            include_context "a8s-pg", :include_shared => true

            context "delete cluster", :order => :defined do
              it "delete the a8s cluster cluster", :clusterop => true do
                skip("Skipping cluster deletion because A9S_E2E_REUSE_CLUSTER=1") if reuse_cluster?
                cmd = "a9s delete cluster a8s -p minikube --verbose --yes"

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
