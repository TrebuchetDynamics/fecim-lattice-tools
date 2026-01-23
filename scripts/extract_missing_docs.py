import json
import os

json_path = "<local-path> EDA suite project progress update_2026-01-23T16-03-20-198Z.json"
output_dir = "<local-path>"

with open(json_path, 'r') as f:
    data = json.load(f)

messages = data.get("chat_messages", [])

created_files = []

for msg in messages:
    attachments = msg.get("attachments", [])
    for att in attachments:
        fname = att.get("file_name", "")
        size = att.get("file_size", 0)
        content = att.get("extracted_content", "")
        
        target_path = ""
        
        if fname == "":
            content_stripped = content.strip()
            # Identify by size or content
            if content_stripped.startswith("# EDA Research Meta-Study"):
                target_path = os.path.join(output_dir, "eda.research.meta-study.md")
            elif content_stripped.startswith("# The Open-Source EDA Ecosystem"):
                target_path = os.path.join(output_dir, "eda.opensource.ecosystem.md")
            elif "EDA Explained Like I'm 5" in content_stripped: # Using 'in' for this one just in case of formatting, but title is unique enough
                target_path = os.path.join(output_dir, "eda.eli5.md")
            else:
                continue
                
            print(f"Writing {target_path} (Size: {size})")
            with open(target_path, 'w') as out:
                out.write(content)
            created_files.append(target_path)

print("Done.")
