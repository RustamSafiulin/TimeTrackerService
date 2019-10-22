
db.createUser(
 {
   user:"test_admin",
   pwd: "qwer~123",
   roles: [
     { role:"readWrite", db:"time_tracker_db" },
     { role:"userAdminAnyDatabase", db:"admin" },
     { role:"dbAdminAnyDatabase", db:"admin" },
     { role:"readWriteAnyDatabase", db:"admin" }
   ]	
 }
)
