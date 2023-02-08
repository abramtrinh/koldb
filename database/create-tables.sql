CREATE DATABASE IF NOT EXISTS koldb;
USE koldb;
-- reminder to remove. using this to just get fresh start.
-- bootstrap or do something on start
DROP TABLE IF EXISTS prices;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS item;

-- CREATE DATABASE IF NOT EXISTS `e_store`
-- CREATE TABLE `e_store`.`brands`( 
-- CONSTRAINT `brand_id` FOREIGN KEY(`brand_id`) REFERENCES `e_store`.`brands`(`id`) ON DELETE RESTRICT ON UPDATE CASCADE ,

CREATE TABLE item (
    itemID INT NOT NULL,
    -- itemName can be empty string. populate with itemID instead if null.
    itemName VARCHAR(40),
    CONSTRAINT item_pk PRIMARY KEY(itemID)
);

CREATE TABLE transactions (
    transID INT NOT NULL AUTO_INCREMENT,
    itemID INT NOT NULL,
    volume INT NOT NULL,
    cost DECIMAL(11,2) NOT NULL,
    epochTime INT NOT NULL,
    CONSTRAINT transactions_pk PRIMARY KEY(transID),
    -- should I be cascading this? any better way to handle it?
    CONSTRAINT transactions_fk FOREIGN KEY (itemID) REFERENCES item(itemID) ON DELETE CASCADE
);

CREATE TABLE prices (
    itemID INT NOT NULL,
    cost INT NOT NULL,
    epochTime INT NOT NULL,
    CONSTRAINT prices_pk PRIMARY KEY(itemID),
    CONSTRAINT prices_fk FOREIGN KEY (itemID) REFERENCES item(itemID) ON DELETE CASCADE
);
