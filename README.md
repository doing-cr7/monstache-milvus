# monstache
a go daemon that syncs MongoDB to Elasticsearch and milvus in realtime. you know, for search. Based on original Monstache by Ryan Wynn
https://github.com/rwynn/monstache




### check user and role

use <logic db>
db.createUser(
  {
    user: "",
    pwd: "",
    roles: [ { role: "readWrite", db: "<logic db>" }]
  }
)

use admin
db.createUser({
  user: "",
  pwd: "",
  roles: [
    { role: "readWrite", db: "admin" },
    { role: "readWrite", db: "<logic db>" },
    { role: "readWrite", db: "monstache" },
    { role: "clusterMonitor", db: "admin" }
  ]
})