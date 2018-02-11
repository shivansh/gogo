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
	sw $t1, v1		# spilled v1, freed $t1
	move $t1, $v0
	li $v0, 1
	move $a0, $t1
	syscall

	# Store variables back into memory
	sw $t1, v3
	sw $t4, v2
	li $v0, 10
	syscall
	.end main
temp:
	li $t1, 1		# v1 -> $t1

	# Store variables back into memory
	sw $t1, v1
