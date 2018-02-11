	.data
v1:	.word	0
v2:	.word	0
v3:	.word	0

	.text


	.globl main
	.ent main
main:
	li $t1, -1		# v1 -> $t1
	li $t4, 2		# v2 -> $t4
	li $v0, 5
	syscall
	move $t3, $v0
	jal temp
	li $v0, 1
	move $a0, $t1
	syscall

	# Store variables back into memory
	sw $t1, v1
	sw $t4, v2
	sw $t3, v3
	li $v0, 10
	syscall
	.end main

	.globl temp
	.ent temp
temp:
	li $t1, 1		# v1 -> $t1
	li $v0, 1
	move $a0, $t1
	syscall

	# Store variables back into memory
	sw $t1, v1

	jr $ra
	.end temp
