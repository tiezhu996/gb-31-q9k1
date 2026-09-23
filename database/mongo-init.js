// MongoDB 初始化脚本（compose 已通过 MONGO_INITDB_ROOT_* 环境变量完成等价初始化）
db.getSiblingDB('petsocial_db').createUser({
  user: 'petsocial_user',
  pwd: 'petsocial_pwd',
  roles: [{ role: 'readWrite', db: 'petsocial_db' }],
});
