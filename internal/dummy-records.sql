-- user data
INSERT INTO users (user, email, password) VALUES
('Jeff77', 'jeff@mail.com', 'Jeff1234'),
('jim24', 'jim@mail.com', 'Jimmy1234'),
('von24', 'von@mail.com', 'Vonna1234');

-- thread data
INSERT INTO threads (title, aid) VALUES
('The Art of Mindful Living', 1),
('Trends in Renewable Energy', 1),
('Exploring Virtual Reality in Education', 1),
('Traveling on a Budget: Tips and Tricks', 1),
('The Science Behind Nutrition Myths', 1),
('Crafting Compelling Short Stories', 2),

-- message data
INSERT INTO messages (tid, aid, body) VALUES
