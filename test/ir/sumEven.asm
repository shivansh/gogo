# Test to find sum of all even numbers less than n

	.data
nStr:	.asciiz "Enter n: "
n:	.word	0
i:	.word	0
k:	.word	0
l:	.word	0
str:	.asciiz "Sum of all even numbers less than n: "

	.text


	.globl main
	.ent main
main:
	li $v0, 4
	la $a0, nStr
	syscall
	li $v0, 5
	syscall
	move $t1, $v0
	li $t4, 0		# i -> $t4
	sw $t1, n		# spilled n, freed $t1
	li $t1, 0		# k -> $t1
	# Store variables back into memory
	sw $t1, k
	sw $t4, i

loop:

	lw $t1, i
	lw $t4, n
	bge $t1, $t4, exit		# exit -> $t0
	rem $t3, $t1, 2		# l -> $t3
	# Store variables back into memory
	sw $t1, i
	sw $t4, n
	sw $t3, l

	lw $t1, l
	beq $t1, 1, skip		# skip -> $t0
	lw $t4, k
	sw $t1, l		# spilled l, freed $t1
	lw $t1, i
	add $t4, $t4, $t1	# k -> $t4
	addi $t1, $t1, 1		# i -> $t1
	# Store variables back into memory
	sw $t1, i
	sw $t4, k

	j loop

skip:
	lw $t1, i
	addi $t1, $t1, 1		# i -> $t1
	# Store variables back into memory
	sw $t1, i

	j loop

exit:
	li $v0, 4
	la $a0, str
	syscall
	li $v0, 1
	lw $t1, k
	move $a0, $t1
	syscall
	# Store variables back into memory
	sw $t1, k
	li $v0, 10
	syscall
	.end main
