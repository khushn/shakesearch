# ShakeSearch

Welcome to the Pulley Shakesearch Take-home Challenge! In this repository,
you'll find a simple web app that allows a user to search for a text string in
the complete works of Shakespeare.

You can see a live version of the app at
https://pulley-shakesearch.herokuapp.com/. Try searching for "Hamlet" to display
a set of results.

In it's current state, however, the app is just a rough prototype. The search is
case sensitive, the results are difficult to read, and the search is limited to
exact matches.

## Your Mission

Improve the search backend. Think about the problem from the **user's perspective**
and prioritize your changes according to what you think is most useful. 

## Evaluation

We will be primarily evaluating based on how well the search works for users. A search result with a lot of features (i.e. multi-words and mis-spellings handled), but with results that are hard to read would not be a strong submission. 


## Submission

1. Fork this repository and send us a link to your fork after pushing your changes. 
2. Heroku hosting - The project includes a Heroku Procfile and, in its
current state, can be deployed easily on Heroku's free tier.
3. In your submission, share with us what changes you made and how you would prioritize changes if you had more time.

## What was added by Me

Link to deployed Heroku app: https://khush-shakesearch.herokuapp.com/
Sample searches to try: 
* Edward thy son, that now is Prince of Wales
* Sonnet
* Hamlet
* Try any text from within the separate titles. It should bring out the paragraph having that phrase, and the title (on the left). The para is sometimes big. 

Main stuff

1. Title reading and indexing - For displaying in searches, if search text from within a title
2. Case insensitive search - SONNET or sonnet. 
3. Paragrapge indexing - Rather than showing -250 to +250. I show that entire paragraph. Sometimes gets too big
4. If Search text is from outside title, limit results and doesn't show book title. 
5. Results are boxed back in JSON. 
6. So had to change the React JS to extract two fields.
7. All the indexing navigation is done in Log(N) time, so pretty decent, even for paragraph indexing

## Future work - enhancements that can be done

1. Indexing could be stored in some persistent storage e.g. Solr or equivalent, if data gets big. May not be an issue here. 
2. Parsing and indexing of Acts in a play, to provide back info on which Act a given search query belongs to
3. Also perhaps thing like which character says a given query. e.g. if its a quoted text from Antonio
4. Remove more clutter from the search results. The returned para get too big often. Take a suitable subset
5. Suggest (auto complete) queries. Can build a lookup, which can prompt for appropriate search queries. 
6. Spelling error tolerance. Can do a lookup based on Soundex or another suitable algo

