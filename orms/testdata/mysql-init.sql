-- The integration test creates and drops databases (AutoCreate/Reset), which
-- needs global DDL privileges on top of the per-database grant the image gives
-- MYSQL_USER.
GRANT ALL PRIVILEGES ON *.* TO 'natools'@'%';
FLUSH PRIVILEGES;
