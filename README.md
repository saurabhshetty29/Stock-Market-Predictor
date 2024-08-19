# Stock-Market-Predictor

Stock-Market-Predictor is a project designed to harness the vast amounts of unstructured data from financial news sources and social media platforms like Reddit. This project aims to aggregate and analyze the discussions and opinions expressed about stocks across various companies, providing real-time sentiment analysis. The tool categorizes sentiments into positive, negative, and neutral. We calculate the ICI score from both news and social media data. The ICI score is a measure of the overall sentiment of the stock which is calculated as:

`ln((1 +positive sentiment count) / (1- negative sentiment count))`

Using this ICI score, we calculate the pearson correlation coefficient between the ICI score and the stock price. This correlation coefficient is used to determine if the stock price follows the sentiment of the news or social media data.
