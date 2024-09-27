CREATE TABLE IF NOT EXISTS songs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255),
    artist VARCHAR(255),
    album VARCHAR(255),
    artwork VARCHAR(255),
    duration VARCHAR(50),
    price VARCHAR(10),
    origin VARCHAR(50)
);

