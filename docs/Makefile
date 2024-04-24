config_staging: guard-PASSWORD
	cd build && echo "force_https: true" > Staticfile
	cd build && htpasswd -nb anynines ${PASSWORD} > Staticfile.auth

deploy_staging:
	cf t -o anynines -s staging
	cd build && cf push a9s-cli-v2-staging

cleanup:
	rm -rf build/Staticfile
	rm -rf build/Staticfile.auth

staging: config_staging deploy_staging cleanup

guard-%:
	@ if [ "${${*}}" = "" ]; then \
        echo "Environment variable $* not set"; \
        exit 1; \
  fi
