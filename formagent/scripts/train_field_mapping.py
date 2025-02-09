from transformers import T5Tokenizer, T5ForConditionalGeneration, Trainer, TrainingArguments

# Load Data
data = [
    {"input": "name, email, resume", "output": "fill(name), fill(email), upload(resume)"},
    {"input": "profile_link, experience_summary, resume", "output": "fill(profile_link), fill(experience_summary), upload(resume)"}
]

# Tokenize
tokenizer = T5Tokenizer.from_pretrained("t5-small")
inputs = tokenizer([d["input"] for d in data], truncation=True, padding=True, return_tensors="pt")
outputs = tokenizer([d["output"] for d in data], truncation=True, padding=True, return_tensors="pt")

# Model
model = T5ForConditionalGeneration.from_pretrained("t5-small")

# Training Arguments
training_args = TrainingArguments(
    output_dir="../models/field_mapper",
    num_train_epochs=3,
    per_device_train_batch_size=8,
    save_total_limit=1,
)

trainer = Trainer(
    model=model,
    args=training_args,
    train_dataset=torch.utils.data.TensorDataset(inputs["input_ids"], outputs["input_ids"]),
)

trainer.train()
model.save_pretrained("../models/field_mapper")
