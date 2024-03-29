ansible-playbook -i inventory.ini java.yml -e ansible_user=guobin -e ansible_port=2222 -e ansible_ssh_private_key_file=~/.ssh/id_rsa -e ansible_become_password=111111
