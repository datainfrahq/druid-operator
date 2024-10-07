#!/bin/sh

set -e

echo "CHECKING SIZE OF CREATED LOOKUP"
LOOKUP_SIZE=$(curl -s http://druid-tiny-cluster-coordinators.druid.svc:8088/druid/coordinator/v1/lookups/config/all | jq '.__default.country_code_names.lookupExtractorFactory.map | length' -r);
echo "LOOKUP_SIZE IS $LOOKUP_SIZE"
if [ $LOOKUP_SIZE != "249" ]
then
  echo "EXPECTED LOOKUP TO CONTAIN 249 KEYS"
  exit 1
else
  echo "SUCCESSFULLY CREATED LOOKUP"
fi

echo "CHECKING THAT LOOKUP LOADS"
for i in {1..100}
do
    sleep 6
    LOADED=$(curl -s http://localhost:8088/druid/coordinator/v1/lookups/status | jq '.__default.country_code_names.loaded' -r);
    if  [ "$LOADED" == "true" ]
    then
        echo "LOOKUP LOADED"
        break
    else
      echo "LOOKUP STILL LOADING..."
    fi
    if [ $i == 100 ]
    then
      echo "LOOKUP DID NOT LOAD WITHIN TIMEOUT"
      exit 1
    fi
done

echo "QUERYING DATA ... "
echo "RUNNING QUERY SELECT LOOKUP('SE', 'country_code_names') AS COUNTRY"

cat > query.json <<EOF
{"query":"SELECT LOOKUP('SE', 'country_code_names') AS COUNTRY","resultFormat":"objectlines"}
EOF

country=`curl -s -XPOST -H'Content-Type: application/json' http://druid-tiny-cluster-routers.druid.svc:8088/druid/v2/sql -d @query.json| jq '.COUNTRY' -r`
echo "COUNTRY IS $country"
if [ $country != "Sweden" ]
then
    echo "QUERY FAILED !!!"
    exit 1
else
    echo "QUERY SUCCESSFUL !!!"
fi
