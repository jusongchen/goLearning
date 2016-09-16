
drop TABLE books
go

CREATE TABLE books (
  id numeric(10) NOT NULL identity,
  title nvarchar(255) NOT NULL,
  author nvarchar(255) NOT NULL,
  description nvarchar(2000) DEFAULT NULL,
  PRIMARY KEY (id)
) 


/*
CREATE TABLE books (
  id numeric(10) NOT NULL identity,
  title varchar(255) NOT NULL,
  author varchar(255) NOT NULL,
  description varchar(2000) DEFAULT NULL,
  PRIMARY KEY (id)
) 
*/

set identity_insert books on

INSERT INTO books (id,title,author,description) VALUES 
(1,'JerBear goes to the City','Garnee Smashington','A young hipster bear seeks his fortune in the wild city of Irvine.')
,(2,'Swarley''s Big Day','Barney Stinson','Putting his Playbook aside, one man seeks a lifetime of happiness.'),
(3,'All Around the Roundabound','Anakin Groundsitter','The riveting tale of a young lad taking pod-racing lessons from an instructor with a dark secret.'),
(4,'Mastering Crossfire: You''ll get caught up in it','Freddie Wong','It''s sometime in the future, the ultimate challenge...  Crossfire!'),
(5,'Time and space','Jusong Uang','A book you must read'),
(6,'Positive Psychology','Tal Ben-Shahar','not a book, but many lecture videos'),
(7,N'红楼梦',N'曹雪芹',N'中国古典文学四大名著之一'),
(8,'The hunger games','Suzanne Collins','The Hunger Games is a 2008 science fiction novel by the American writer Suzanne Collins. It is written in the voice of 16-year-old Katniss Everdeen, who lives in the dystopian, post-apocalyptic nation of Panem in North America. The Capitol, a highly advanced metropolis, exercises political control over the rest of the nation. The Hunger Games are an annual event in which one boy and one girl aged 12–18 from each of the twelve districts surrounding the Capitol are selected by lottery to compete in a televised battle to the death.'),(9,'Harry Potter','J. K. Rowling','Harry Potter is a series of seven fantasy novels written by the British author J. K. Rowling. The series, named after the titular character, chronicles the adventures of a young wizard, Harry Potter, and his friends Ronald Weasley and Hermione Granger, all of whom are students at Hogwarts School of Witchcraft and Wizardry. The main story arc concerns Harry''s quest to overcome the Dark wizard Lord Voldemort, who aims to become immortal, conquer the wizarding world, subjugate non-magical people, and destroy all those who stand in his way, especially Harry Potter.'),(10,'Divergent','Veronica Roth','A story.'),(11,'Gone with the Wind ','Margaret Mitchell','Gone with the Wind is a novel written by Margaret Mitchell, first published in 1936. The story is set in Clayton County, Georgia, and Atlanta during the American Civil War and Reconstruction. It depicts the experiences of Scarlett O''Hara, the spoiled daughter of a well-to-do plantation owner, who must use every means at her disposal to come out of the poverty she finds herself in after Sherman''s \"March to the Sea\". A historical novel, the story is a Bildungsroman or coming-of-age story, with the novel''s title taken from a poem written by the British poet, Ernest Dowson'),(12,'Divergent','the autheor','new york best seller');



select * from books