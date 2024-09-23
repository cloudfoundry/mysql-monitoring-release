require 'rspec'
require 'json'
require 'yaml'
require 'bosh/template/test'

describe 'jobs/mysql-diag/config/mysql-diag-config.yml' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '..')) }
  let(:job) { release.job('mysql-diag') }
  let(:template) { job.template('config/mysql-diag-config.yml') }
  let(:spec) { { "db_username" => "mysql-diag-user", "db_password" => "mysql-diag-password" } }
  let(:mysql_link) {
    Bosh::Template::Test::Link.new(
      name: 'mysql',
      instances: [Bosh::Template::Test::LinkInstance.new(id: 'mysql0.uuid', name: 'mysql', address: 'mysql0.address')],
      properties: { "port" => 6033 },
    )
  }
  let(:diag_agent_link) {
    Bosh::Template::Test::Link.new(
      name: 'mysql-diag-agent',
      instances: [Bosh::Template::Test::LinkInstance.new(address: 'mysql-diag-agent-dns')],
      properties: {
        "mysql-monitoring" => {
          "mysql-diag-agent" => {
            "username" => "diag-agent-user",
            "password" => "diag-agent-password",
            "port" => 8765,
            "tls" => { "enabled" => true, "ca" => "mysql-diag-agent-ca", "server_name" => "mysql-diag-agent-identity" }
          },
        },
      },
    )
  }

  let(:links) { [mysql_link, diag_agent_link,] }
  let(:rendered_template) { template.render(spec, consumes: links) }
  let(:parsed_config) { YAML.load(rendered_template) }

  it 'renders' do
    expect { rendered_template }.to_not raise_error
  end

  it 'includes configuration for mysql' do
    expect(parsed_config.keys).to match_array(%w[mysql])
  end

  it 'configures the database credentials' do
    expect(parsed_config).to include("mysql" => hash_including(
      "username" => "mysql-diag-user",
      "password" => "mysql-diag-password",
    ))
  end

  it 'configures the mysql port exported by the mysql job' do
    expect(parsed_config).to include("mysql" => hash_including(
      "port" => 6033,
    ))
  end

  it 'configures information to communicate with the mysql-diag-agent' do
    expect(parsed_config).to include("mysql" => hash_including(
      "agent" => {
        "username" => "diag-agent-user",
        "password" => "diag-agent-password",
        "port" => 8765,
        "tls" => { "enabled" => true, "ca" => "mysql-diag-agent-ca", "server_name" => "mysql-diag-agent-identity" }
      }
    ))
  end

  it 'sets hardcoded thresholds for warning about disk utilization' do
    expect(parsed_config).to include("mysql" => hash_including(
      "threshold" => {
        "disk_used_warning_percent" => 80,
        "disk_inodes_used_warning_percent" => 80,
      }
    ))
  end

  it 'configures information for accessing the set of mysql instances' do
    expect(parsed_config).to include("mysql" => hash_including(
      "nodes" => [{ "host" => "mysql0.address", "name" => "mysql", "uuid" => "mysql0.uuid" }]
    ))
  end

  context 'required db credentials are not provided' do
    let(:spec) { { } }
    it 'errors on rendering' do
       expect { rendered_template }.to raise_error 'No database credentials were configured. db_username and db_password properties must be specified for the mysql-diag job.'
    end
  end

end
