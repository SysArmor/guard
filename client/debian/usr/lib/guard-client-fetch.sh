#!/bin/bash
/usr/bin/guard-client ca --address=${ADDRESS} --node-id=${NODE_ID} --node-secret=${NODE_SECRET}
/usr/bin/guard-client principals --address=${ADDRESS} --node-id=${NODE_ID} --node-secret=${NODE_SECRET}
/usr/bin/guard-client revoke-keys --address=${ADDRESS} --node-id=${NODE_ID} --node-secret=${NODE_SECRET}
