CREATE KEYSPACE demo
WITH replication = {
  'class': 'SimpleStrategy',
  'replication_factor': 1
};

USE demo;

CREATE TABLE logs (
  id UUID PRIMARY KEY,
  text text,
  created_at timestamp
);