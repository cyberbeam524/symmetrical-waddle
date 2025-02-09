#     curl https://api-inference.huggingface.co/models/deepset/roberta-base-squad2 \
#         -X POST \
#         -d `{"inputs": { "question": "What are repeated containers that can be extracted with css selectors from this html page?", "context": """HTML page: <!DOCTYPE html>
#     <html lang="en">
#     <head>
#         <meta charset="UTF-8">
#         <meta name="viewport" content="width=device-width, initial-scale=1.0">
#         <title>Product Listings</title>
#         <style>
#             .product-card {
#                 border: 1px solid #ccc;
#                 padding: 16px;
#                 margin: 16px 0;
#                 display: flex;
#                 flex-direction: column;
#                 align-items: center;
#             }
#             .product-title { font-size: 1.5em; font-weight: bold; }
#             .product-price { color: green; font-weight: bold; }
#             .product-description { font-size: 0.9em; margin: 8px 0; }
#         </style>
#     </head>
#     <body>
#         <h1>Product Listings</h1>
#         <div class="product-container">
#             <!-- Product 1 -->
#             <div class="product-card">
#                 <img src="product1.jpg" alt="Product 1" class="product-image">
#                 <h2 class="product-title">Product 1</h2>
#                 <p class="product-price">$19.99</p>
#                 <p class="product-description">This is the first product description.</p>
#                 <a href="/product/1" class="product-link">View Details</a>
#             </div>

#             <!-- Product 2 -->
#             <div class="product-card">
#                 <img src="product2.jpg" alt="Product 2" class="product-image">
#                 <h2 class="product-title">Product 2</h2>
#                 <p class="product-price">$29.99</p>
#                 <p class="product-description">This is the second product description.</p>
#                 <a href="/product/2" class="product-link">View Details</a>
#             </div>

#             <!-- Product 3 -->
#             <div class="product-card">
#                 <img src="product3.jpg" alt="Product 3" class="product-image">
#                 <h2 class="product-title">Product 3</h2>
#                 <p class="product-price">$39.99</p>
#                 <p class="product-description">This is the third product description.</p>
#                 <a href="/product/3" class="product-link">View Details</a>
#             </div>
#         </div>
#     </body>
#     </html>
#     """ }}` \
#         -H 'Content-Type: application/json' \
#         -H 'Authorization: Bearer hf'

# curl https://api-inference.huggingface.co/models/deepset/roberta-base-squad2 \
#     -X POST \
#     -d '{"inputs": {"question": "What is the css selectors to scrape repeated containers from this html page?", "context": "HTML page: <!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n    <title>Product Listings</title>\n    <style>\n        .product-card {\n            border: 1px solid #ccc;\n            padding: 16px;\n            margin: 16px 0;\n            display: flex;\n            flex-direction: column;\n            align-items: center;\n        }\n        .product-title { font-size: 1.5em; font-weight: bold; }\n        .product-price { color: green; font-weight: bold; }\n        .product-description { font-size: 0.9em; margin: 8px 0; }\n    </style>\n</head>\n<body>\n    <h1>Product Listings</h1>\n    <div class=\"product-container\">\n        <!-- Product 1 -->\n        <div class=\"product-card\">\n            <img src=\"product1.jpg\" alt=\"Product 1\" class=\"product-image\">\n            <h2 class=\"product-title\">Product 1</h2>\n            <p class=\"product-price\">$19.99</p>\n            <p class=\"product-description\">This is the first product description.</p>\n            <a href=\"/product/1\" class=\"product-link\">View Details</a>\n        </div>\n\n        <!-- Product 2 -->\n        <div class=\"product-card\">\n            <img src=\"product2.jpg\" alt=\"Product 2\" class=\"product-image\">\n            <h2 class=\"product-title\">Product 2</h2>\n            <p class=\"product-price\">$29.99</p>\n            <p class=\"product-description\">This is the second product description.</p>\n            <a href=\"/product/2\" class=\"product-link\">View Details</a>\n        </div>\n\n        <!-- Product 3 -->\n        <div class=\"product-card\">\n            <img src=\"product3.jpg\" alt=\"Product 3\" class=\"product-image\">\n            <h2 class=\"product-title\">Product 3</h2>\n            <p class=\"product-price\">$39.99</p>\n            <p class=\"product-description\">This is the third product description.</p>\n            <a href=\"/product/3\" class=\"product-link\">View Details</a>\n        </div>\n    </div>\n</body>\n</html>"}}' \
#     -H 'Content-Type: application/json' \
#     -H 'Authorization: Bearer hfsecret'


from huggingface_hub import InferenceClient

client = InferenceClient(api_key="")

messages = [
	{
		"role": "user",
		"content": "What is the css selectors to scrape repeated containers from this html page? HTML page: <!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n    <title>Product Listings</title>\n    <style>\n        .product-card {\n            border: 1px solid #ccc;\n            padding: 16px;\n            margin: 16px 0;\n            display: flex;\n            flex-direction: column;\n            align-items: center;\n        }\n        .product-title { font-size: 1.5em; font-weight: bold; }\n        .product-price { color: green; font-weight: bold; }\n        .product-description { font-size: 0.9em; margin: 8px 0; }\n    </style>\n</head>\n<body>\n    <h1>Product Listings</h1>\n    <div class=\"product-container\">\n        <!-- Product 1 -->\n        <div class=\"product-card\">\n            <img src=\"product1.jpg\" alt=\"Product 1\" class=\"product-image\">\n            <h2 class=\"product-title\">Product 1</h2>\n            <p class=\"product-price\">$19.99</p>\n            <p class=\"product-description\">This is the first product description.</p>\n            <a href=\"/product/1\" class=\"product-link\">View Details</a>\n        </div>\n\n        <!-- Product 2 -->\n        <div class=\"product-card\">\n            <img src=\"product2.jpg\" alt=\"Product 2\" class=\"product-image\">\n            <h2 class=\"product-title\">Product 2</h2>\n            <p class=\"product-price\">$29.99</p>\n            <p class=\"product-description\">This is the second product description.</p>\n            <a href=\"/product/2\" class=\"product-link\">View Details</a>\n        </div>\n\n        <!-- Product 3 -->\n        <div class=\"product-card\">\n            <img src=\"product3.jpg\" alt=\"Product 3\" class=\"product-image\">\n            <h2 class=\"product-title\">Product 3</h2>\n            <p class=\"product-price\">$39.99</p>\n            <p class=\"product-description\">This is the third product description.</p>\n            <a href=\"/product/3\" class=\"product-link\">View Details</a>\n        </div>\n    </div>\n</body>\n</html>"
	}
]

completion = client.chat.completions.create(
    model="Qwen/Qwen2.5-Coder-32B-Instruct", 
	messages=messages, 
	max_tokens=500
)

print(completion.choices[0].message)