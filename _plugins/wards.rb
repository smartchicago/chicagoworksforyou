module Jekyll

  class WardPage < Page
    def initialize(site, base, ward)
      @site = site
      @base = base
      @dir = File.join('wards', ward)
      @name = 'index.html'

      self.process(@name)
      self.read_yaml(File.join(base, '_layouts'), 'ward.html')
      self.data['ward'] = ward
      self.data['title'] = "Ward #{ward}"
    end
  end

  class WardPageGenerator < Generator
    safe true

    def generate(site)
      if site.layouts.key? 'ward'
        dir = site.config['category_dir'] || 'wards'
        for i in 1..50 do
          site.pages << WardPage.new(site, site.source, i.to_s)
        end
      end
    end
  end

end
