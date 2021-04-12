start transaction;

create table journal (
    id bigserial primary key,
    time timestamp not null
);

create index on journal(time);

create type transport as enum ('tcp', 'udp');

create table connection (
    id serial primary key,
    transport transport not null default 'udp',
    community text not null default '',
    timeout int not null default 5,
    retries int not null default 2,
    max_oids int not null default 60,
    max_repetitions int not null default 25,
    msg_flags int not null default 3, /* see https://pkg.go.dev/github.com/gosnmp/gosnmp#SnmpV3MsgFlags */
    security_model int not null default 3, /* see https://pkg.go.dev/github.com/gosnmp/gosnmp#SnmpV3SecurityModel */
    auth_protocol int not null default 3, /* see https://pkg.go.dev/github.com/gosnmp/gosnmp#SnmpV3AuthProtocol */
    username text not null,
    password text not null,
    priv_protocol int not null default 3, /* see https://pkg.go.dev/github.com/gosnmp/gosnmp#SnmpV3PrivProtocol */
    priv_password text not null
);

create table hostname (
    id bigserial primary key,
    hostname text constraint unique_hostname unique not null
);

create table system (
    id bigserial primary key,
    name text constraint unique_system_name unique not null,
    hostname_id bigint references hostname(id),
    port int not null default 161,
    connection_id int references connection(id)
);

create index on system(hostname_id);

create table mac_address (
    id bigserial primary key,
    mac_address text constraint unique_mac_address unique not null
);

create function port_number(name text) returns int[3] as
	$$ select regexp_matches(name, '(\d)+/(\d+)/(\d+)')::int[3] $$
language sql immutable;

create table port (
    id bigserial primary key,
    system_id bigint not null references system(id),
    mac_address_id bigint not null references mac_address(id),
    name text not null,
    number int[3] generated always as (port_number(name)) stored,
    description text not null,
    constraint unique_port_system_name unique(system_id, name)
);

create index on port(system_id);
create index on port(mac_address_id);

create table port_journal (
    journal_id bigint not null references journal(id),
    port_id bigint not null references port(id),
    status text not null,
    speed int not null
);

create index on port_journal(journal_id);
create index on port_journal(port_id);

create table lldp (
    id bigserial primary key,
    local_port_id bigint not null references port(id),
    remote_port_id bigint not null references port(id),
    constraint unique_lldp unique(local_port_id, remote_port_id)
);

create index on lldp(local_port_id);
create index on lldp(remote_port_id);

create table lldp_journal (
    journal_id bigint not null references journal(id),
    lldp_id bigint not null references lldp(id)
);

create index on lldp_journal(journal_id);
create index on lldp_journal(lldp_id);

create table mac_address_journal (
    journal_id bigint not null references journal(id),
    mac_address_id bigint not null references mac_address(id),
    port_id bigint not null references port(id),
    vlan int not null
);

create index on mac_address_journal(journal_id);
create index on mac_address_journal(mac_address_id);
create index on mac_address_journal(port_id);

create table ip_address (
    id bigserial primary key,
    ip_address text constraint unique_ip_address unique not null
);

create table arp (
    id bigserial primary key,
    mac_address_id bigint not null references mac_address(id),
    ip_address_id bigint not null references ip_address(id),
    constraint unique_arp unique(mac_address_id, ip_address_id)
);

create index on arp(mac_address_id);
create index on arp(ip_address_id);

create table arp_journal (
    journal_id bigint not null references journal(id),
    arp_id bigint not null references arp(id)
);

create index on arp_journal(journal_id);
create index on arp_journal(arp_id);

create table resolve (
    id bigserial primary key,
    ip_address_id bigint not null references ip_address(id),
    hostname_id bigint not null references hostname(id),
    constraint unique_resolve unique(ip_address_id, hostname_id)
);

create index on resolve(ip_address_id);
create index on resolve(hostname_id);

create table resolve_journal (
    journal_id bigint not null references journal(id),
    resolve_id bigint not null references resolve(id)
);

create index on resolve_journal(journal_id);
create index on resolve_journal(resolve_id);

create table vendor (
    prefix text primary key,
    name text not null
);

create view mac_address_vendor as
    select id as mac_address_id, prefix as vendor_prefix from mac_address join vendor on
        substring(mac_address from 1 for 8) = prefix or
        substring(mac_address from 1 for 11) = prefix or
        substring(mac_address from 1 for 14) = prefix
;

end transaction;
