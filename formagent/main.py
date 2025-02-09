import asyncio
from transformers import BertForSequenceClassification, BertTokenizer, T5ForConditionalGeneration, T5Tokenizer
from playwright.async_api import async_playwright

# Load Models
classifier = BertForSequenceClassification.from_pretrained("models/classifier")
classifier_tokenizer = BertTokenizer.from_pretrained("bert-base-uncased")

mapper = T5ForConditionalGeneration.from_pretrained("models/field_mapper")
mapper_tokenizer = T5Tokenizer.from_pretrained("t5-small")

async def automate_form(url):
    async with async_playwright() as p:
        browser = await p.chromium.launch()
        page = await browser.new_page()
        await page.goto(url)

        # Extract form
        html = await page.content()
        inputs = classifier_tokenizer(html, return_tensors="pt", truncation=True, padding=True)
        pred = classifier(inputs["input_ids"])
        form_type = "JobStreet" if pred.logits.argmax() == 0 else "LinkedIn"

        print(f"Detected Form Type: {form_type}")

        # Map Fields to Actions
        field_input = mapper_tokenizer("name, email, resume", return_tensors="pt")
        actions = mapper.generate(field_input["input_ids"])
        print(f"Recommended Actions: {mapper_tokenizer.decode(actions[0])}")

        await browser.close()

# Test with a sample URL
asyncio.run(automate_form("https://example.com/job_form"))
