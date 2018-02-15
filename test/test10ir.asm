# Test to demonstrate function call

	.data
n:	.word	0
i:	.word	0
first:	.word	0
second:	.word	0
return:	.word	0
temp:	.word	0

	.text


	.globl main
	.ent main
main:
	li $v0, 5
	syscall
	move $t1, $v0
	li $t4, 0		# i -> $t4
	sw $t1, n		# spilled n, freed $t1
	li $t1, 0		# first -> $t1
	sw $t4, i		# spilled i, freed $t4
	li $t4, 1		# second -> $t4
	sw $t1, first
	sw $t4, second
	jal fib
	lw $t1, first
	lw $t4, second
	sw $t1, first		# spilled first, freed $t1
	move $t1, $v0

	# Store variables back into memory
	sw $t1, return
	sw $t4, second
	li $v0, 10
	syscall
	.end main

	.globl fib
	.ent fib
fib:
	li $v0, 10
	syscall
	.end main
loop:
	li $v0, 10
	syscall
	.end main
	lw $t1, i
	lw $t4, n
	bgt $t1, $t4, exit		# exit -> $t0
	lw $t3, second
	sw $t4, n		# spilled n, freed $t4
	move $t4, $t3		# temp -> $t4
	lw $t2, first
	add $t3, $t3, $t2	# second -> $t3
	move $t2, $t4		# first -> $t2
	addi $t1, $t1, 1		# i -> $t1

	# Store variables back into memory
	sw $t1, i
	sw $t4, temp
	sw $t3, second
	sw $t2, first
	li $v0, 10
	syscall
	.end main
	j loop
	li $v0, 10
	syscall
	.end main
exit:
	li $v0, 1
	lw $t1, second
	move $a0, $t1
	syscall
	move $v0, $t1

	# Store variables back into memory
	sw $t1, second

	jr $ra
	.end fib
