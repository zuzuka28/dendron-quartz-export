# dendron-quartz-export

This is a utility for exporting notes [dendron](https://www.dendron.so) in [quartz](https://quartz.jzhao .xyz)

# Features

- converting dot-notation to a file hierarchy
- mapping fields from frontmatter
- conversion of dendron links and syntax to obsidian flavored markdown

# Config Example

```yaml
dendron_notes_path: .my_notes # vautl with notes
export_path: .public_notes # directory to which the result will be uploaded
frontmatter_replace_field:
  # renaming the frontmatter desc field to description
  - field: desc
    replace: description

  # formatting the list of tags in OFM
  - field: tags
    field_type: tags
    replace: tags

  # formatting the list of links in OFM
  - field: links
    field_type: links
    replace: links

  # formatting the list of links in OFM
  - field: opinionFor
    field_type: links
    replace: opinionFor

  # converting timestamp from created field to date in date field
  - field: created
    field_type: timestamp
    replace: date
```
