# from transformers import AutoModelForCausalLM, AutoTokenizer
# import torch

# # Load the model and tokenizer
# model_name = "Qwen/Qwen2.5-Coder-32B-Instruct"  # Replace with your desired model
# tokenizer = AutoTokenizer.from_pretrained(model_name)
# model = AutoModelForCausalLM.from_pretrained(model_name)

# # Move model to GPU (if available)
# device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
# model.to(device)

# # Example inference function
# def generate_response(prompt):
#     inputs = tokenizer(prompt, return_tensors="pt").to(device)
#     outputs = model.generate(inputs["input_ids"], max_length=50, do_sample=True, temperature=0.7)
#     response = tokenizer.decode(outputs[0], skip_special_tokens=True)
#     return response

# part1 v2:
from transformers import AutoModelForCausalLM, AutoTokenizer
import torch
model_name = "Qwen/Qwen2.5-Coder-32B-Instruct"

model = AutoModelForCausalLM.from_pretrained(
    model_name,
    torch_dtype="auto",
    device_map="auto"
)
tokenizer = AutoTokenizer.from_pretrained(model_name)

prompt = "write a quick sort algorithm."
messages = [
    {"role": "system", "content": "You are Qwen, created by Alibaba Cloud. You are a helpful assistant."},
    {"role": "user", "content": prompt}
]
text = tokenizer.apply_chat_template(
    messages,
    tokenize=False,
    add_generation_prompt=True
)
model_inputs = tokenizer([text], return_tensors="pt").to(model.device)

generated_ids = model.generate(
    **model_inputs,
    max_new_tokens=512
)
generated_ids = [
    output_ids[len(input_ids):] for input_ids, output_ids in zip(model_inputs.input_ids, generated_ids)
]

response = tokenizer.batch_decode(generated_ids, skip_special_tokens=True)[0]

print(f"response:{response}")


# Test the inference function

# messages = [
# 	{
# 		"role": "user",
# 		"content": "What is the css selectors to scrape repeated containers from this html page? HTML page: <!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n    <title>Product Listings</title>\n    <style>\n        .product-card {\n            border: 1px solid #ccc;\n            padding: 16px;\n            margin: 16px 0;\n            display: flex;\n            flex-direction: column;\n            align-items: center;\n        }\n        .product-title { font-size: 1.5em; font-weight: bold; }\n        .product-price { color: green; font-weight: bold; }\n        .product-description { font-size: 0.9em; margin: 8px 0; }\n    </style>\n</head>\n<body>\n    <h1>Product Listings</h1>\n    <div class=\"product-container\">\n        <!-- Product 1 -->\n        <div class=\"product-card\">\n            <img src=\"product1.jpg\" alt=\"Product 1\" class=\"product-image\">\n            <h2 class=\"product-title\">Product 1</h2>\n            <p class=\"product-price\">$19.99</p>\n            <p class=\"product-description\">This is the first product description.</p>\n            <a href=\"/product/1\" class=\"product-link\">View Details</a>\n        </div>\n\n        <!-- Product 2 -->\n        <div class=\"product-card\">\n            <img src=\"product2.jpg\" alt=\"Product 2\" class=\"product-image\">\n            <h2 class=\"product-title\">Product 2</h2>\n            <p class=\"product-price\">$29.99</p>\n            <p class=\"product-description\">This is the second product description.</p>\n            <a href=\"/product/2\" class=\"product-link\">View Details</a>\n        </div>\n\n        <!-- Product 3 -->\n        <div class=\"product-card\">\n            <img src=\"product3.jpg\" alt=\"Product 3\" class=\"product-image\">\n            <h2 class=\"product-title\">Product 3</h2>\n            <p class=\"product-price\">$39.99</p>\n            <p class=\"product-description\">This is the third product description.</p>\n            <a href=\"/product/3\" class=\"product-link\">View Details</a>\n        </div>\n    </div>\n</body>\n</html>"
# 	}
# ]

# completion = client.chat.completions.create(
#     model="Qwen/Qwen2.5-Coder-32B-Instruct", 
# 	messages=messages, 
# 	max_tokens=500
# )

prompt = "What is the css selectors to scrape repeated containers from this html page? HTML page: <!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n    <title>Product Listings</title>\n    <style>\n        .product-card {\n            border: 1px solid #ccc;\n            padding: 16px;\n            margin: 16px 0;\n            display: flex;\n            flex-direction: column;\n            align-items: center;\n        }\n        .product-title { font-size: 1.5em; font-weight: bold; }\n        .product-price { color: green; font-weight: bold; }\n        .product-description { font-size: 0.9em; margin: 8px 0; }\n    </style>\n</head>\n<body>\n    <h1>Product Listings</h1>\n    <div class=\"product-container\">\n        <!-- Product 1 -->\n        <div class=\"product-card\">\n            <img src=\"product1.jpg\" alt=\"Product 1\" class=\"product-image\">\n            <h2 class=\"product-title\">Product 1</h2>\n            <p class=\"product-price\">$19.99</p>\n            <p class=\"product-description\">This is the first product description.</p>\n            <a href=\"/product/1\" class=\"product-link\">View Details</a>\n        </div>\n\n        <!-- Product 2 -->\n        <div class=\"product-card\">\n            <img src=\"product2.jpg\" alt=\"Product 2\" class=\"product-image\">\n            <h2 class=\"product-title\">Product 2</h2>\n            <p class=\"product-price\">$29.99</p>\n            <p class=\"product-description\">This is the second product description.</p>\n            <a href=\"/product/2\" class=\"product-link\">View Details</a>\n        </div>\n\n        <!-- Product 3 -->\n        <div class=\"product-card\">\n            <img src=\"product3.jpg\" alt=\"Product 3\" class=\"product-image\">\n            <h2 class=\"product-title\">Product 3</h2>\n            <p class=\"product-price\">$39.99</p>\n            <p class=\"product-description\">This is the third product description.</p>\n            <a href=\"/product/3\" class=\"product-link\">View Details</a>\n        </div>\n    </div>\n</body>\n</html>"
response = generate_response(prompt)
print(f"Model Response: {response}")



# # part2 
# feedback_data = [
#     {"input": "What is the capital of France?", "model_output": "Paris", "correct_output": "Paris"},
#     {"input": "Tell me about the Eiffel Tower.", "model_output": "It is in London.", "correct_output": "It is in Paris."},
# ]

# # Save feedback to a JSON or CSV file
# import json
# with open("feedback.json", "w") as f:
#     json.dump(feedback_data, f)


# # part3
# from datasets import Dataset

# # Load feedback data
# with open("feedback.json") as f:
#     feedback_data = json.load(f)

# # Prepare dataset
# def preprocess_data(entry):
#     return {
#         "input_ids": tokenizer(entry["input"], truncation=True, padding="max_length", max_length=512)["input_ids"],
#         "labels": tokenizer(entry["correct_output"], truncation=True, padding="max_length", max_length=512)["input_ids"],
#     }

# dataset = Dataset.from_list(feedback_data).map(preprocess_data)


# # part4:
# from transformers import TrainingArguments, Trainer

# training_args = TrainingArguments(
#     output_dir="fine_tuned_model",
#     per_device_train_batch_size=4,
#     num_train_epochs=2,
#     learning_rate=5e-5,
#     logging_dir="logs",
#     save_strategy="epoch",
#     save_total_limit=2,
#     fp16=torch.cuda.is_available(),  # Use mixed precision if available
# )

# trainer = Trainer(
#     model=model,
#     args=training_args,
#     train_dataset=dataset,
#     tokenizer=tokenizer,
# )

# # Fine-tune the model
# trainer.train()

# # Save the fine-tuned model
# model.save_pretrained("fine_tuned_model")
# tokenizer.save_pretrained("fine_tuned_model")
