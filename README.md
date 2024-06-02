Full-Text engine implementation in Go

Full-Text Search (FTS) is a technique for searching text in a collection of documents. A document can refer to a web page, a newspaper article, an email message, or any structured text. Most well-known FTS engine is Elasticsearch.

<h2>Explanation of the implementation</h2>
- <h3>Step 1 : Document Definition</h3>
Define the document structure and the global variables for documents and indexes.

- <h3>Step 2 : Initialize Inverted and TF-IDF Indexes</h3>
Initialize the inverted and TF-IDF indexes by calling the init function when the program starts.

- <h3>Step 3 : Tokenize Text</h3>
Create a function to tokenize the text into words and normalize them to lowercase.

- <h3>Step 4 : Calculate Term Frequency (TF)</h3>
Define a function to calculate the term frequency for the tokens in a document.

- <h3>Step 5 : Build Inverted Index</h3>
Build an inverted index that maps each term to the list of document IDs that contain the term.

- <h3>Step 6 : Calculate Inverse Document Frequency (IDF)</h3>
Calculate the IDF for each term based on the inverted index and the total number of documents.

- <h3>Step 7 : Build TF-IDF Index</h3>
Build the TF-IDF index by combining the term frequency and inverse document frequency.

- <h3>Step 8 : Perform TF-IDF Search</h3>
Perform a TF-IDF search on the query and return a map of document IDs to their TF-IDF scores.

- <h3>Step 9 : Perform Letter-by-Letter Search</h3>
Perform a secondary letter-by-letter search for suggestions based on the query.

- <h3>Step 10 : Rank Search Results</h3>
Rank the search results based on their TF-IDF scores.

- <h3>Step 11 : Handle Search Requests</h3>
Handle the search requests by combining the results from both TF-IDF and letter-by-letter searches.

- <h3>Step 12 : Serve Index Page and Start Server</h3>
Serve the index page and start the HTTP server.