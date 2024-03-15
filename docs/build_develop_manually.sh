export VERSION=$(cat ../anynines-deployment/CHANGELOG.md | grep '## \[' | egrep -oe '([0-9]+(\.[0-9]+)+)' | head -n 1)
echo $VERSION

cp -r ../anynines-deployment/docs/* docs

ruby ./convert_anynines_deployment_changelog.rb
ruby ./build_versions_json.rb
yarn run docusaurus build
