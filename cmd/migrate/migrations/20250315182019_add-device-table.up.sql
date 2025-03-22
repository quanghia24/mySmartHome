CREATE TABLE IF NOT EXISTS devices (
    `feedId` INT UNSIGNED NOT NULL,
    `feedKey` VARCHAR(255) NOT NULL,
    `title` VARCHAR(255) NOT NULL,
    `userId` INT UNSIGNED NOT NULL,
    `roomId` INT UNSIGNED NOT NULL,
    
    PRIMARY KEY(`feedId`),
    FOREIGN KEY(`userId`) REFERENCES users(`id`),
    FOREIGN KEY(`roomId`) REFERENCES rooms(`id`)
)