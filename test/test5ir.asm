	.data
temp:	.space	16

	.text


	.globl main
	.ent main
main:
	la $t4, temp
	li $t1, 4		# temp[2] -> $t1
	addi $t3, $t1, 2	# temp[1] -> $t3
	li $v0, 1
	move $a0, $t3
	syscall

	# Store variables back into memory
	sw $t1, 8($t4)
	sw $t3, 4($t4)
	li $v0, 10
	syscall
	.end main
