# Table: github_languages

The primary key for this table is **_cq_id**.

## Columns

| Name          | Type          |
| ------------- | ------------- |
|_cq_id (PK)|`uuid`|
|_cq_parent_id|`uuid`|
|full_name|`utf8`|
|name|`utf8`|
|languages|`list<item: utf8, nullable>`|