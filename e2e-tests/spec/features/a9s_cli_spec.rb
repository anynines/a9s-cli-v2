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

      safe_output = output.encode("UTF-8", invalid: :replace, undef: :replace, replace: "")
      expect(safe_output).to include("The #{@service_instance_name} system appears to be ready. All expected pods are running.")
      kubectl_verify_pg_service_instance_exists(@service_instance_name, @workload_namespace)
      kubectl_verify_pg_pods_running(@service_instance_name, @workload_namespace, 3)
      master_service_name = "#{@service_instance_name}-master"
      kubectl_verify_service_exists(master_service_name, @workload_namespace)
      kubectl_verify_secret_exists("postgres.credentials.#{@service_instance_name}", @workload_namespace)
      kubectl_verify_secret_exists("standby.credentials.#{@service_instance_name}", @workload_namespace)
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
      kubectl_verify_service_binding_implemented(@service_binding_name, @workload_namespace)

      secret_name = "#{@service_binding_name}-service-binding"
      kubectl_verify_secret_exists(secret_name, @workload_namespace)

      binding_database = kubectl_secret_data(secret_name, @workload_namespace, "database")
      binding_instance_service = kubectl_secret_data(secret_name, @workload_namespace, "instance_service")
      binding_username = kubectl_secret_data(secret_name, @workload_namespace, "username")
      binding_password = kubectl_secret_data(secret_name, @workload_namespace, "password")

      expect(binding_database).not_to be_empty
      expect(binding_instance_service).not_to be_empty
      expect(binding_username).not_to be_empty
      expect(binding_password).not_to be_empty

      service_name = binding_instance_service
      service_namespace = @workload_namespace
      if binding_instance_service.include?(".")
        parts = binding_instance_service.split(".")
        service_name = parts[0]
        service_namespace = parts[1] unless parts[1].to_s.empty?
      end
      kubectl_verify_service_exists(service_name, service_namespace)

      master_pod = kubectl_pg_master_pod_name(@service_instance_name, @workload_namespace)
      psql_output = kubectl_psql_select_one(
        master_pod,
        @workload_namespace,
        binding_instance_service,
        binding_username,
        binding_password,
        binding_database
      )
      expect(psql_output).to match(/\b1\b/)
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
      secret_name = "#{@service_binding_name}-service-binding"
      kubectl_verify_secret_not_exists(secret_name, @workload_namespace)
    end
  end

  context "backups" do
    it "creates a backup of an a8s pg service instance" do
      cmd = "a9s create pg backup --name #{@backup_name} -i #{@service_instance_name} -n #{@workload_namespace} --verbose --yes"

      logger.info(cmd)

      output = `#{cmd}`

      logger.info "\t" + output

      safe_output = output.encode("UTF-8", invalid: :replace, undef: :replace, replace: "")
      expect(safe_output).to include("The backup with the name #{@backup_name} in namespace #{@workload_namespace} has been successful")
      kubectl_verify_pg_backup_exists(@backup_name, @workload_namespace)
      kubectl_verify_pg_condition_true("backups", @backup_name, @workload_namespace, "Complete")
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

      safe_output = output.encode("UTF-8", invalid: :replace, undef: :replace, replace: "")
      expect(safe_output).to include("The restore with the name #{@restore_name} in namespace #{@workload_namespace} has been successful.")
      kubectl_verify_pg_restore_exists(@restore_name, @workload_namespace)
      kubectl_verify_pg_condition_true("restores", @restore_name, @workload_namespace, "Complete")
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
      kubectl_verify_pg_pods_not_exists(@service_instance_name, @workload_namespace, 3)
      master_service_name = "#{@service_instance_name}-master"
      kubectl_verify_service_not_exists(master_service_name, @workload_namespace)
      kubectl_verify_secret_not_exists("postgres.credentials.#{@service_instance_name}", @workload_namespace)
      kubectl_verify_secret_not_exists("standby.credentials.#{@service_instance_name}", @workload_namespace)
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
            kubectl_verify_crd_exists("postgresqls.postgresql.anynines.com")
            kubectl_verify_crd_exists("servicebindings.servicebindings.anynines.com")
            kubectl_verify_crd_exists("backups.backups.anynines.com")
            kubectl_verify_crd_exists("restores.backups.anynines.com")
            kubectl_verify_pods_running_by_label("a8s-system", "app.kubernetes.io/name=backup-manager")
            kubectl_verify_pods_running_by_label("a8s-system", "app.kubernetes.io/name=postgresql-controller-manager")
            kubectl_verify_pods_running_by_label("a8s-system", "app.kubernetes.io/name=service-binding-controller-manager")
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
            kubectl_verify_crd_exists("postgresqls.postgresql.anynines.com")
            kubectl_verify_crd_exists("servicebindings.servicebindings.anynines.com")
            kubectl_verify_crd_exists("backups.backups.anynines.com")
            kubectl_verify_crd_exists("restores.backups.anynines.com")
            kubectl_verify_pods_running_by_label("a8s-system", "app.kubernetes.io/name=backup-manager")
            kubectl_verify_pods_running_by_label("a8s-system", "app.kubernetes.io/name=postgresql-controller-manager")
            kubectl_verify_pods_running_by_label("a8s-system", "app.kubernetes.io/name=service-binding-controller-manager")
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
              kubectl_verify_crd_exists("postgresqls.postgresql.anynines.com")
              kubectl_verify_crd_exists("servicebindings.servicebindings.anynines.com")
              kubectl_verify_crd_exists("backups.backups.anynines.com")
              kubectl_verify_crd_exists("restores.backups.anynines.com")
              kubectl_verify_pods_running_by_label("a8s-system", "app.kubernetes.io/name=backup-manager")
              kubectl_verify_pods_running_by_label("a8s-system", "app.kubernetes.io/name=postgresql-controller-manager")
              kubectl_verify_pods_running_by_label("a8s-system", "app.kubernetes.io/name=service-binding-controller-manager")
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
              kubectl_verify_crd_exists("postgresqls.postgresql.anynines.com")
              kubectl_verify_crd_exists("servicebindings.servicebindings.anynines.com")
              kubectl_verify_crd_exists("backups.backups.anynines.com")
              kubectl_verify_crd_exists("restores.backups.anynines.com")
              kubectl_verify_pods_running_by_label("a8s-system", "app.kubernetes.io/name=backup-manager")
              kubectl_verify_pods_running_by_label("a8s-system", "app.kubernetes.io/name=postgresql-controller-manager")
              kubectl_verify_pods_running_by_label("a8s-system", "app.kubernetes.io/name=service-binding-controller-manager")
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
      context "klutch", order: :defined do
        context "control-plane", :aws => true, :very_expensive => true do
          before(:context) do
            unless aws_cli_available?
              skip("Skipping Klutch control-plane tests. Requires aws_cli.")
            end
            unless aws_credentials_available?
              skip("Skipping Klutch control-plane tests. Requires logged-in aws_cli.")
            end

            @klutch_hosted_zone = ENV.fetch("A9S_E2E_KLUTCH_HOSTED_ZONE", "hub.test.a9s.io")
            suffix = Time.now.utc.strftime("%H%M%S")
            @workload_namespace = "a8s-workload"
            @klutch_service_instance_name = "klutch-pg-#{suffix}"
            @klutch_service_binding_name = "#{@klutch_service_instance_name}-sb"
          end

          it "creates a Klutch control plane cluster (VERY EXPENSIVE; requires aws_cli)", :clusterop => true, :slow => true, :very_expensive => true do
            cmd = "a9s create cluster klutch control-plane -p aws --verbose --yes"
            cmd += " --hosted-zone-name #{@klutch_hosted_zone}"
            cmd += " --tenant-operator-bind-url https://klutch-bind.#{@klutch_hosted_zone}/bind-noninteractive"

            logger.info cmd
            output = `#{cmd}`
            logger.info "\t" + output

            expect($?.success?).to be(true)
            expect(output).to include("Klutch control plane EKS cluster is ready.")

            region = klutch_control_plane_region
            vpc_id = aws_verify_vpc_exists(
              { "Klutch" => "ControlPlane", "Name" => "klutch-control-plane-vpc" },
              region: region,
              timeout_seconds: 900
            )
            expect(vpc_id).not_to eq("")
            expect(aws_eks_cluster_exists?("klutch-control-plane", region: region)).to be(true)
          end

          it "creates a Klutch tenant (VERY EXPENSIVE; requires existing control plane)", :clusterop => true, :slow => true, :very_expensive => true do
            tenant_name = klutch_e2e_tenant_name
            expect(tenant_name).not_to eq("")

            tenant_namespace = "a9s-tenants-operator-system"
            kubectl_verify_crd_exists("tenants.klutch.anynines.com")

            cmd = "a9s create klutch tenant --tenant-name #{tenant_name} --yes"
            logger.info cmd

            output = `#{cmd}`
            logger.info "\t" + output

            expect($?.success?).to be(true)
            expect(output).to include("Tenant #{tenant_name} created via tenant operator")

            kubectl_wait_for_tenant_ready(tenant_name, namespace: tenant_namespace, timeout_seconds: 900)

            region = klutch_control_plane_region
            user_pool_id = aws_verify_cognito_user_pool_exists({ "Klutch" => "ControlPlane" }, region: region, timeout_seconds: 900)
            expect(user_pool_id).not_to eq("")

            secret_name = "klutch/#{tenant_name}/oidc-client"
            aws_verify_secretsmanager_secret_exists(secret_name, region: region, timeout_seconds: 900)
            aws_verify_tenant_oidc_secret_valid(secret_name, region: region, timeout_seconds: 900)

            secret_json = aws_secretsmanager_get_secret_json(secret_name, region: region)
            expect(secret_json["bind_url"].to_s).to include("klutch-bind.#{@klutch_hosted_zone}")
          end

          it "creates a Klutch workload cluster with 1 EKS node (VERY EXPENSIVE; requires existing control plane and tenant)", :clusterop => true, :slow => true, :very_expensive => true do
            tenant_name = klutch_e2e_tenant_name
            expect(tenant_name).not_to eq("")

            region = klutch_control_plane_region
            unless aws_eks_cluster_exists?("klutch-control-plane", region: region)
              skip("Skipping Klutch workload test. Requires existing EKS cluster klutch-control-plane in region #{region}.")
            end

            secret_name = "klutch/#{tenant_name}/oidc-client"
            unless aws_secretsmanager_secret_exists?(secret_name, region: region)
              skip("Skipping Klutch workload test. Requires tenant secret #{secret_name} in Secrets Manager (region #{region}).")
            end
            begin
              aws_verify_tenant_oidc_secret_valid(secret_name, region: region, timeout_seconds: 180)
            rescue RuntimeError => e
              skip("Skipping Klutch workload test. Tenant secret #{secret_name} is not ready: #{e.message}")
            end

            cmd = "a9s create cluster klutch workload -p aws --tenant-name #{tenant_name} --eks-nodes 1 --yes"
            logger.info cmd

            output = `#{cmd}`
            logger.info "\t" + output

            workload_cluster_name = extract_klutch_workload_cluster_name(output)
            klutch_e2e_set_workload_cluster_name(workload_cluster_name) unless workload_cluster_name.to_s.strip == ""

            expect($?.success?).to be(true)
            expect(output).to include("Klutch workload EKS cluster is ready.")
            expect(output).to include("Workload cluster bound to control plane using tenant secret.")

            expect(workload_cluster_name).not_to eq("")
            expect(aws_eks_cluster_exists?(workload_cluster_name, region: region)).to be(true)

            nodegroup_name = "#{workload_cluster_name}-nodegroup"
            scaling = aws_eks_nodegroup_scaling_config(workload_cluster_name, nodegroup_name, region: region)
            expect(scaling["min"]).to eq(1)
            expect(scaling["max"]).to eq(1)
            expect(scaling["desired"]).to eq(1)
          end

          it "creates a PostgreSQL service instance on the Klutch workload cluster", :clusterop => true, :slow => true, :very_expensive => true do
            region = klutch_control_plane_region
            workload_cluster_name = klutch_e2e_workload_cluster_name
            if workload_cluster_name.to_s.strip == ""
              skip("Skipping service instance test. No workload cluster name stored (run workload create first).")
            end
            unless aws_eks_cluster_exists?(workload_cluster_name, region: region)
              skip("Skipping service instance test. EKS cluster #{workload_cluster_name} not found in region #{region}.")
            end

            aws_update_kubeconfig(workload_cluster_name, region: region)
            kubectl_verify_context_contains(workload_cluster_name)
            kubectl_ensure_namespace(@workload_namespace)
            kubectl_wait_for_crd_exists("postgresqlinstances.anynines.com", timeout_seconds: 600)
            kubectl_wait_for_crd_exists("servicebindings.anynines.com", timeout_seconds: 600)

            cmd = "a9s create klutch pg instance --name #{@klutch_service_instance_name} -n #{@workload_namespace} --verbose --yes"
            logger.info cmd
            output = `#{cmd}`
            logger.info "\t" + output

            expect($?.success?).to be(true)
            expect(output).to include("Klutch PostgreSQL instance #{@klutch_service_instance_name} created in namespace #{@workload_namespace}.")

            kubectl_verify_klutch_pg_service_instance_exists(@klutch_service_instance_name, @workload_namespace)
            kubectl_verify_pg_condition_true("postgresqlinstances.anynines.com", @klutch_service_instance_name, @workload_namespace, "Ready")

            aws_update_kubeconfig("klutch-control-plane", region: region)
            kubectl_verify_context_contains("klutch-control-plane")
            provider_namespace = kubectl_find_resource_namespace("postgresqlinstances.anynines.com", @klutch_service_instance_name)
            expect(provider_namespace).not_to eq("")
            @klutch_provider_namespace = provider_namespace
            kubectl_verify_pg_condition_true("postgresqlinstances.anynines.com", @klutch_service_instance_name, provider_namespace, "Ready")

            aws_update_kubeconfig(workload_cluster_name, region: region)
            kubectl_verify_context_contains(workload_cluster_name)
          end

          it "creates a PostgreSQL service binding on the Klutch workload cluster", :clusterop => true, :slow => true, :very_expensive => true do
            region = klutch_control_plane_region
            workload_cluster_name = klutch_e2e_workload_cluster_name
            if workload_cluster_name.to_s.strip == ""
              skip("Skipping service binding test. No workload cluster name stored (run workload create first).")
            end
            unless aws_eks_cluster_exists?(workload_cluster_name, region: region)
              skip("Skipping service binding test. EKS cluster #{workload_cluster_name} not found in region #{region}.")
            end

            aws_update_kubeconfig(workload_cluster_name, region: region)
            kubectl_verify_context_contains(workload_cluster_name)
            kubectl_wait_for_crd_exists("servicebindings.anynines.com", timeout_seconds: 600)
            unless kubectl_klutch_pg_service_instance_exists?(@klutch_service_instance_name, @workload_namespace)
              skip("Skipping service binding test. Service instance #{@klutch_service_instance_name} not found in namespace #{@workload_namespace}.")
            end

            cmd = "a9s create klutch pg servicebinding --name #{@klutch_service_binding_name} --service-instance #{@klutch_service_instance_name} -n #{@workload_namespace} --verbose --yes"
            logger.info cmd
            output = `#{cmd}`
            logger.info "\t" + output

            expect($?.success?).to be(true)
            expect(output).to include("Klutch PostgreSQL service binding #{@klutch_service_binding_name} created in namespace #{@workload_namespace}.")

            kubectl_verify_klutch_service_binding_exists(@klutch_service_binding_name, @workload_namespace)
            kubectl_verify_klutch_service_binding_implemented(@klutch_service_binding_name, @workload_namespace)
            secret_name = "#{@klutch_service_binding_name}-service-binding"
            kubectl_verify_secret_exists(secret_name, @workload_namespace)

            binding_database = kubectl_secret_data(secret_name, @workload_namespace, "database")
            binding_instance_service = kubectl_secret_data(secret_name, @workload_namespace, "instance_service")
            binding_username = kubectl_secret_data(secret_name, @workload_namespace, "username")
            binding_password = kubectl_secret_data(secret_name, @workload_namespace, "password")
            expect(binding_database).not_to be_empty
            expect(binding_instance_service).to include("#{@klutch_service_instance_name}-master")
            expect(binding_username).not_to be_empty
            expect(binding_password).not_to be_empty

            aws_update_kubeconfig("klutch-control-plane", region: region)
            kubectl_verify_context_contains("klutch-control-plane")
            provider_namespace = @klutch_provider_namespace.to_s
            if provider_namespace == ""
              provider_namespace = kubectl_find_resource_namespace("servicebindings.anynines.com", @klutch_service_binding_name)
              @klutch_provider_namespace = provider_namespace unless provider_namespace.to_s == ""
            end
            expect(provider_namespace).not_to eq("")
            kubectl_verify_klutch_service_binding_exists(@klutch_service_binding_name, provider_namespace)
            kubectl_verify_klutch_service_binding_implemented(@klutch_service_binding_name, provider_namespace)

            aws_update_kubeconfig(workload_cluster_name, region: region)
            kubectl_verify_context_contains(workload_cluster_name)
          end

          it "deletes the PostgreSQL service binding from the Klutch workload cluster", :clusterop => true, :slow => true, :very_expensive => true do
            region = klutch_control_plane_region
            workload_cluster_name = klutch_e2e_workload_cluster_name
            if workload_cluster_name.to_s.strip == ""
              skip("Skipping service binding deletion test. No workload cluster name stored (run workload create first).")
            end
            unless aws_eks_cluster_exists?(workload_cluster_name, region: region)
              skip("Skipping service binding deletion test. EKS cluster #{workload_cluster_name} not found in region #{region}.")
            end

            aws_update_kubeconfig(workload_cluster_name, region: region)
            kubectl_verify_context_contains(workload_cluster_name)
            unless kubectl_klutch_service_binding_exists?(@klutch_service_binding_name, @workload_namespace)
              skip("Skipping service binding deletion test. Service binding #{@klutch_service_binding_name} not found in namespace #{@workload_namespace}.")
            end

            cmd = "a9s delete klutch pg servicebinding --name #{@klutch_service_binding_name} -n #{@workload_namespace} --verbose --yes"
            logger.info cmd
            output = `#{cmd}`
            logger.info "\t" + output

            expect($?.success?).to be(true)
            expect(output).to include("Klutch service binding #{@klutch_service_binding_name} successfully deleted from namespace #{@workload_namespace}.")

            kubectl_verify_klutch_service_binding_not_exists(@klutch_service_binding_name, @workload_namespace)

            provider_namespace = @klutch_provider_namespace.to_s
            if provider_namespace != ""
              aws_update_kubeconfig("klutch-control-plane", region: region)
              kubectl_verify_context_contains("klutch-control-plane")
              kubectl_verify_klutch_service_binding_not_exists(@klutch_service_binding_name, provider_namespace)
              aws_update_kubeconfig(workload_cluster_name, region: region)
              kubectl_verify_context_contains(workload_cluster_name)
            end
          end

          it "deletes the PostgreSQL service instance from the Klutch workload cluster", :clusterop => true, :slow => true, :very_expensive => true do
            region = klutch_control_plane_region
            workload_cluster_name = klutch_e2e_workload_cluster_name
            if workload_cluster_name.to_s.strip == ""
              skip("Skipping service instance deletion test. No workload cluster name stored (run workload create first).")
            end
            unless aws_eks_cluster_exists?(workload_cluster_name, region: region)
              skip("Skipping service instance deletion test. EKS cluster #{workload_cluster_name} not found in region #{region}.")
            end

            aws_update_kubeconfig(workload_cluster_name, region: region)
            kubectl_verify_context_contains(workload_cluster_name)
            unless kubectl_klutch_pg_service_instance_exists?(@klutch_service_instance_name, @workload_namespace)
              skip("Skipping service instance deletion test. Service instance #{@klutch_service_instance_name} not found in namespace #{@workload_namespace}.")
            end

            cmd = "a9s delete klutch pg instance --name #{@klutch_service_instance_name} -n #{@workload_namespace} --verbose --yes"
            logger.info cmd
            output = `#{cmd}`
            logger.info "\t" + output

            expect($?.success?).to be(true)
            expect(output).to include("Klutch service instance #{@klutch_service_instance_name} successfully deleted from namespace #{@workload_namespace}.")

            kubectl_verify_klutch_pg_service_instance_not_exists(@klutch_service_instance_name, @workload_namespace)
            kubectl_verify_secret_not_exists("#{@klutch_service_binding_name}-service-binding", @workload_namespace, timeout_seconds: 300)

            provider_namespace = @klutch_provider_namespace.to_s
            if provider_namespace != ""
              aws_update_kubeconfig("klutch-control-plane", region: region)
              kubectl_verify_context_contains("klutch-control-plane")
              kubectl_verify_klutch_pg_service_instance_not_exists(@klutch_service_instance_name, provider_namespace)
              aws_update_kubeconfig(workload_cluster_name, region: region)
              kubectl_verify_context_contains(workload_cluster_name)
            end
          end

          it "deletes the Klutch workload cluster (VERY EXPENSIVE; cleanup)", :clusterop => true, :slow => true, :very_expensive => true, :cleanup => true do
            region = klutch_control_plane_region

            workload_cluster_name = klutch_e2e_workload_cluster_name
            if workload_cluster_name.to_s.strip == ""
              skip("Skipping Klutch workload cleanup. No workload cluster name stored (run create first).")
            end
            unless aws_eks_cluster_exists?(workload_cluster_name, region: region)
              skip("Skipping Klutch workload cleanup. EKS cluster #{workload_cluster_name} not found in region #{region}.")
            end

            vpc_id = aws_find_vpc_id_by_tags(
              { "Klutch" => "Workload", "Name" => "#{workload_cluster_name}-vpc" },
              region: region
            )
            expect(vpc_id).not_to eq("")

            cmd = "a9s delete cluster klutch workload -p aws --cluster-name #{workload_cluster_name} --yes --really"
            logger.info cmd

            output = `#{cmd}`
            logger.info "\t" + output

            expect($?.success?).to be(true)
            expect(output).to include("Cluster deleted.")
            aws_verify_eks_cluster_deleted(workload_cluster_name, region: region, timeout_seconds: 1200)
            expect(aws_eks_cluster_exists?(workload_cluster_name, region: region)).to be(false)
            aws_verify_vpc_deleted(vpc_id, region: region, timeout_seconds: 900)
            expect(
              aws_find_vpc_id_by_tags({ "Klutch" => "Workload", "Name" => "#{workload_cluster_name}-vpc" }, region: region)
            ).to eq("")
          end

          it "deletes the Klutch control plane cluster (VERY EXPENSIVE; requires aws_cli)", :clusterop => true, :slow => true, :very_expensive => true, :destroy_control_plane => true do
            region = klutch_control_plane_region
            vpc_id = aws_find_vpc_id_by_tags(
              { "Klutch" => "ControlPlane", "Name" => "klutch-control-plane-vpc" },
              region: region
            )
            expect(vpc_id).not_to eq("")

            cmd = "a9s delete cluster klutch control-plane -p aws --verbose --yes --really"
            logger.info cmd

            output = `#{cmd}`
            logger.info "\t" + output

            expect($?.success?).to be(true)
            expect(output).to include("Cluster deleted.")

            aws_verify_eks_cluster_deleted("klutch-control-plane", region: region, timeout_seconds: 1200)
            expect(aws_eks_cluster_exists?("klutch-control-plane", region: region)).to be(false)
            aws_verify_vpc_deleted(vpc_id, region: region, timeout_seconds: 900)
            expect(
              aws_find_vpc_id_by_tags({ "Klutch" => "ControlPlane", "Name" => "klutch-control-plane-vpc" }, region: region)
            ).to eq("")
          end
        end
      end
    end
  end
end
