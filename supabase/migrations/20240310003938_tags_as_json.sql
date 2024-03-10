-- Step 1: Add a new column with the desired data type
ALTER TABLE collectable_instances ADD COLUMN tags_jsonb jsonb;

-- Step 2: Migrate data from the old column to the new column
UPDATE collectable_instances SET tags_jsonb = array_to_json(tags)::jsonb;

-- Step 3: Remove the old column
ALTER TABLE collectable_instances DROP COLUMN tags;

-- Step 4: Rename the new column to the original column name
ALTER TABLE collectable_instances RENAME COLUMN tags_jsonb TO tags;
