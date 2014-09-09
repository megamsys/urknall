# encoding: utf-8

class NavLinker

  def initialize(item)
    @item = item
  end

  def nav_link(item, path)
    link_text = item[:short_title] || item[:title]

    html_class = html_class_for(item)

    li(a(link_text, :href => path), :class => html_class)
  end

  private

  def html_class_for(item)
    html_classes = []

    if @item == item || (item.identifier != '/' && @item.identifier.start_with?(item.identifier))
      html_classes << 'active'
    end

    html_classes.empty? ? nil : html_classes.join(' ')
  end

  def a(content, params = {})
    href = params.fetch(:href)

    start_tag = %(<a href="#{href}">)
    end_tag   = '</a>'

    start_tag + content + end_tag
  end

  def span(content)
    start_tag = '<span>'
    end_tag   = '</span>'

    start_tag + content + end_tag
  end

  def li(content, params = {})
    html_class = params.fetch(:class, nil)

    start_tag = html_class ? %(<li class="#{html_class}">) : '<li>'
    end_tag   = '</li>'

    start_tag + content + end_tag
  end

end

module NavLinkHelper

  def nav_link(item)
    NavLinker.new(@item).nav_link(item, relative_path_to(item))
  end

end

include NavLinkHelper