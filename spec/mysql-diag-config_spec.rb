require 'rspec'
require 'json'
require 'yaml'
require 'bosh/template/test'

describe 'jobs/mysql-diag/config/mysql-diag-config.yml' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '..')) }
  let(:job) { release.job('mysql-diag') }
  let(:template) { job.template('config/mysql-diag-config.yml') }
  let(:spec) { {} }
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
  let(:canary_link) {
    Bosh::Template::Test::Link.new(
      name: 'replication-canary',
      instances: [Bosh::Template::Test::LinkInstance.new(address: 'replication-canary-dns')],
      properties: {
        "mysql-monitoring" => {
          "replication-canary" => {
            "canary_username" => "replication-canary-basic-auth-username",
            "canary_password" => "replication-canary-basic-auth-password",
            "api_port" => 1234,
            "tls" => { "enabled" => true, "ca" => "replication-canary-ca", "server_name" => "replication-canary-identity" }
          }
        }
      }
    )
  }
  let(:links) { [mysql_link, diag_agent_link, canary_link,] }
  let(:rendered_template) { template.render(spec, consumes: links) }
  let(:parsed_config) { YAML.load(rendered_template) }

  it 'renders' do
    expect { rendered_template }.to_not raise_error
  end

  it 'includes configuration for mysql and replication-canary' do
    expect(parsed_config.keys).to match_array(%w[mysql canary])
  end

  it 'renders configuration to communicate with replication-canary' do
    expect(parsed_config).to include("canary" => {
      "username" => "replication-canary-basic-auth-username",
      "password" => "replication-canary-basic-auth-password",
      "api_port" => 1234,
      "tls" => { "enabled" => true, "ca" => "replication-canary-ca", "server_name" => "replication-canary-identity" }
    })
  end

  it 'configures database credentials exported by the replication-canary job' do
    expect(parsed_config).to include("mysql" => hash_including(
      "username" => "replication-canary-basic-auth-username",
      "password" => "replication-canary-basic-auth-password",
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

  context 'when explicit mysql database credentials are provided' do
    let(:spec) {
      { "db_username" => "mysql-diag-user", "db_password" => "mysql-diag-password" }
    }
    it 'uses the explicit credentials and does not use credentials from replication-canary link' do
      expect(parsed_config).to include("mysql" => hash_including(
        "username" => "mysql-diag-user",
        "password" => "mysql-diag-password",
      ))
    end
  end

  describe 'without replication-canary' do
    let(:links) { [mysql_link, diag_agent_link] }

    it 'errors when no explicit database credentials have been provided' do
      expect { rendered_template }.to raise_error 'No database credentials were configured. db_username and db_password properties must be specified for the mysql-diag job.'
    end

    context 'when only the db_username was provided, but not db_password' do
      let(:spec) {
        { "db_username" => "something" }
      }

      it 'errors because no explicit database password was provided' do
        expect { rendered_template }.to raise_error 'No database credentials were configured. db_username and db_password properties must be specified for the mysql-diag job.'
      end
    end

    context 'when only the db_password was provided, but not db_username' do
      let(:spec) {
        { "db_password" => "something" }
      }

      it 'errors because no explicit database password was provided' do
        expect { rendered_template }.to raise_error 'No database credentials were configured. db_username and db_password properties must be specified for the mysql-diag job.'
      end
    end

    context 'when explicit credentials are provided' do
      let(:spec) {
        { "db_username" => "user-specified-mysql-diag-db-username", "db_password" => "user-specified-mysql-diag-db-password" }
      }

      it 'configures the explicit credentials and does not use credentials from replication-canary link' do
        expect(parsed_config).to include("mysql" => hash_including(
          "username" => "user-specified-mysql-diag-db-username",
          "password" => "user-specified-mysql-diag-db-password",
        ))
      end

      it 'only includes configuration for mysql and omits replication canary config' do
        expect(parsed_config.keys).to eq(%w[mysql])
      end

      it 'still configures the mysql port exported by the mysql job' do
        expect(parsed_config).to include("mysql" => hash_including(
          "port" => 6033,
        ))
      end

      it 'still configures information to communicate with the mysql-diag-agent' do
        expect(parsed_config).to include("mysql" => hash_including(
          "agent" => {
            "username" => "diag-agent-user",
            "password" => "diag-agent-password",
            "port" => 8765,
            "tls" => { "enabled" => true, "ca" => "mysql-diag-agent-ca", "server_name" => "mysql-diag-agent-identity" }
          }
        ))
      end

      it 'still sets hardcoded thresholds for warning about disk utilization' do
        expect(parsed_config).to include("mysql" => hash_including(
          "threshold" => {
            "disk_used_warning_percent" => 80,
            "disk_inodes_used_warning_percent" => 80,
          }
        ))
      end

      it 'still configures information for accessing the set of mysql instances' do
        expect(parsed_config).to include("mysql" => hash_including(
          "nodes" => [{ "host" => "mysql0.address", "name" => "mysql", "uuid" => "mysql0.uuid" }]
        ))
      end
    end
  end
end
