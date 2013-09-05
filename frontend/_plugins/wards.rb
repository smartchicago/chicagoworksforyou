require 'json'

module Jekyll

  class WardPage < Page
    def initialize(site, base, ward)
      @site = site
      @base = base
      @dir = File.join('ward', ward)
      @name = 'index.html'

      self.process(@name)
      self.read_yaml(File.join(base, '_layouts'), 'ward.html')

      ward_paths = read_data_object(base, 'wards.json')
      self.data['points'] = ward_paths['data'][ward]

      aldermen = read_data_object(base, 'aldermen2.json')
      ward_data = aldermen['data'][ward]
      self.data['alderman'] = ward_data['ALDERMAN']
      self.data['website'] = ward_data['WEBSITE']
      self.data['twitter'] = ward_data['TWITTER']
      self.data['facebook'] = ward_data['FACEBOOK']
      self.data['councilmatic'] = ward_data['COUNCILMATIC']
      self.data['ward'] = ward
      self.data['title'] = "Ward #{ward}"
    end

    def read_data_object(base, filename)
      data_path = File.join(base, 'data')
      if File.symlink?(data_path)
        return "Data directory '#{data_path}' cannot be a symlink"
      end
      file = File.join(data_path, filename)

      return "File #{file} could not be found" if !File.exists?( file )

      result = nil
      Dir.chdir(data_path) do
        result = File.read( filename )
      end
      puts "## Error: No data in #{file}" if result.nil?
      # puts result
      result = JSON.parse( result ) if result
      { 'data' => result,
        'mtime' => File.mtime(file) }
    end

  end

  class WardPageGenerator < Generator
    safe true

    def generate(site)
      if site.layouts.key? 'ward'
        for i in 1..50 do
          site.pages << WardPage.new(site, site.source, i.to_s)
        end
      end
    end
  end

end
