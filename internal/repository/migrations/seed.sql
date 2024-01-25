CREATE DATABASE IF NOT EXISTS `complexus` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
USE `complexus`;
CREATE TABLE IF NOT EXISTS issues (
  id serial PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  description text NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  deleted_at TIMESTAMP NOT NULL
);
CREATE TABLE IF NOT EXISTS projects (
  id serial PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description text NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  deleted_at TIMESTAMP NOT NULL
)