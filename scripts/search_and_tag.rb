def search_and_tag(current_case, search, tag)
    items = current_case.search(search)
    for item in items
        item.add_tag(tag)
        puts("Tagged item: #{item}")
    end
    puts("Tagged #{items.length} items with tag: #{tag}")
end