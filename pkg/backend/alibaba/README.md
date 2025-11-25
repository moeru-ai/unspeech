Scripts used to scrape the `voices.json`:

> Can be directly executed within the browser

```javascript
rows = Array.from(document.querySelector('#eca844883b0ox > tbody').children)

data = rows.map((r) => {
  const colsValues = []
  const columns = Array.from((r.children))
  for (const col of columns) {
    const audio = col.querySelector('audio')
    if (!audio)
      colsValues.push(col.textContent)
    else colsValues.push(audio.src)
  }

  return colsValues
})

cols = [
  { field: 'name' },
  { field: 'preview_audio_url' },
  { field: 'model' },
  { field: 'voice' },
  { field: 'scenarios' },
  { field: 'language' },
  { field: 'bitrate' },
  { field: 'format' }
]

results = data.map((d) => {
  const obj = {}
  for (let i = 0; i < d.length; i++) {
    const field = cols[i].field
    obj[field] = d[i]
  }

  return obj
})
```
