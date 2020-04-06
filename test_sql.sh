#!/bin/bash
export TNS_ADMIN=/function/wallet
cat $TNS_ADMIN/tnsnames.ora
cat $TNS_ADMIN/sqlnet.ora
echo "id=>" `id`
###  replace the below conn string with your own
sqlplus admin/WElcome1412#@adwfree_low << EOF
select * from v\$session;
EOF


