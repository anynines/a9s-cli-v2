export VERSION=$(cat ../anynines-deployment/CHANGELOG.md | grep '## \[' | egrep -oe '([0-9]+(\.[0-9]+)+)' | head -n 1)
echo $VERSION

# NOTE use following code to update old Versions, be aware to cahnge theanynines-deployment repo to the correct old version
DOC_DIR="versioned_docs/version-${VERSION}"
if [ -d $DOC_DIR ]; then
	echo "Version already exists, will copy content only"
else
	echo "Generate new Version in Docusaurus"
	yarn run docusaurus docs:version $VERSION
fi
rm -Rf "${DOC_DIR}/*"
cp -r ../anynines-deployment/docs/* $DOC_DIR

ruby ./convert_anynines_deployment_changelog.rb
ruby ./build_versions_json.rb
yarn run docusaurus build
