package main

import (
    "bufio"
    "context"
    "fmt"
    "github.com/jomei/notionapi"
    "os"
    "time"
)

const (
    NotionAPIToken              string = "NOTION_API_TOKEN"
    ValueUnitDatabaseID                = "VALUE_UNIT_DATABASE_ID"
    ValueUnitGraphicFocus              = "💵 🎧 ⏳ 👨‍💻 👉"
    ValueUnitGraphicDefault            = "=========="
    PropertyTitleValueUnitStart        = "value unit start"
    PropertyTitleValueUnitEnd          = "value unit end"
    PropertyTtitleFocus                = "focus"
)

func main() {

    notionAPIToken := os.Getenv(NotionAPIToken)
    dbID := os.Getenv(ValueUnitDatabaseID)
    client := notionapi.NewClient(notionapi.Token(notionAPIToken))

    ticker := time.NewTicker(1 * time.Minute)

    for curTime := range ticker.C {
        mins := curTime.Minute()
        fmt.Println(mins)
        switch mins {
        case 0, 15, 30, 45:
            fmt.Println("value unit focus update", curTime.String())
            updateFocusUnitInDatabaseOrPanic(client, notionapi.DatabaseID(dbID))
        default:
            fmt.Println("default case")
            continue
        }
    }

}

func updateFocusUnitInDatabaseOrPanic(c *notionapi.Client, dbID notionapi.DatabaseID) {
    db, err := c.Database.Query(context.Background(), dbID, &notionapi.DatabaseQueryRequest{})
    if err != nil {
        panic(err.Error())
    }
    for _, page := range db.Results {
        err := updateValueUnitFocus(context.Background(), c, notionapi.PageID(page.ID))
        if err != nil {
            panic(err.Error())
        }
    }
}

func updateValueUnitFocus(ctx context.Context, c *notionapi.Client, pageId notionapi.PageID) error {
    page, err := c.Page.Get(ctx, pageId)
    if err != nil {
        return err
    }
    // decide default or focus window
    live := isValueUnitLive(page)

    // send update call
    return updateValueUnitFocusGraphic(ctx, c, page, live)

}

func updateValueUnitFocusGraphic(ctx context.Context, c *notionapi.Client, page *notionapi.Page, live bool) error {
    valueUnitGraphic := ValueUnitGraphicDefault
    if live {
        valueUnitGraphic = ValueUnitGraphicFocus
    }
    propertyVal := page.Properties[PropertyTtitleFocus].(*notionapi.RichTextProperty)
    curGraphic := ""
    if len(propertyVal.RichText) > 0 {
        curGraphic = propertyVal.RichText[0].Text.Content
    }
    if curGraphic == valueUnitGraphic {
        return nil
    }
    if len(propertyVal.RichText) == 0 {
        propertyVal.RichText = []notionapi.RichText{
            {
                Type:     notionapi.ObjectTypeText,
                Text:     &notionapi.Text{},
                Mention:  nil,
                Equation: nil,
                Annotations: &notionapi.Annotations{
                    Color: "default",
                },
                PlainText: "",
                Href:      "",
            },
        }
    }
    propertyVal.RichText[0].Text.Content = valueUnitGraphic
    page.Properties[PropertyTtitleFocus] = propertyVal
    _, err := c.Page.Update(ctx, notionapi.PageID(page.ID), &notionapi.PageUpdateRequest{
        Properties: page.Properties,
    })
    if err != nil {
        return err
    }
    return err
}

func isValueUnitLive(p *notionapi.Page) bool {
    valueUnitStart, valueUnitEnd := 0, 0
    live := false
    for title, property := range p.Properties {
        if title != PropertyTitleValueUnitStart && title != PropertyTitleValueUnitEnd {
            continue
        }
        if title == PropertyTitleValueUnitStart {
            valueUnitStart = int(property.(*notionapi.NumberProperty).Number)
        }
        if title == PropertyTitleValueUnitEnd {
            valueUnitEnd = int(property.(*notionapi.NumberProperty).Number)
        }
    }
    currentTimeVal := currentTime()
    if currentTimeVal >= valueUnitStart && currentTimeVal < valueUnitEnd {
        live = true
        fmt.Println("new value unit focus:", valueUnitStart, "to", valueUnitEnd)
    }
    return live
}

func currentTime() int {
    t := time.Now()
    hrs := t.Hour()
    mins := t.Minute()
    return hrs*100 + mins
}

func prompt() {
    scanner := bufio.NewScanner(os.Stdin)
    fmt.Printf("press return to continue...")
    for scanner.Scan() {
        break
    }
    if err := scanner.Err(); err != nil {
        panic(err.Error())
    }
}
