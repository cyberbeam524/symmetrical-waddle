import torch
from transformers import BertTokenizer, BertForSequenceClassification, Trainer, TrainingArguments
from sklearn.model_selection import train_test_split

# Load Data
import json
with open("../data/forms.json") as f:
    data = json.load(f)

texts = [item["html"] for item in data]
labels = [0 if item["website"] == "JobStreet" else 1 for item in data]  # 0: JobStreet, 1: LinkedIn

# Tokenize
tokenizer = BertTokenizer.from_pretrained("bert-base-uncased")
inputs = tokenizer(texts, padding=True, truncation=True, return_tensors="pt")
labels = torch.tensor(labels)

# Split Data
train_texts, val_texts, train_labels, val_labels = train_test_split(inputs["input_ids"], labels, test_size=0.2)

# Model
model = BertForSequenceClassification.from_pretrained("bert-base-uncased", num_labels=2)

# Trainer
training_args = TrainingArguments(
    output_dir="../models/classifier",
    num_train_epochs=3,
    per_device_train_batch_size=8,
    evaluation_strategy="epoch",
    save_total_limit=1,
    logging_dir="../outputs/logs",
)

train_dataset = torch.utils.data.TensorDataset(train_texts, train_labels)
val_dataset = torch.utils.data.TensorDataset(val_texts, val_labels)

trainer = Trainer(
    model=model,
    args=training_args,
    train_dataset=train_dataset,
    eval_dataset=val_dataset,
)

trainer.train()
model.save_pretrained("../models/classifier")
